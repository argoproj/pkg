package s3

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"github.com/minio/minio-go/v7/pkg/sse"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const nullIAMEndpoint = ""

type S3Client interface {
	// PutFile puts a single file to a bucket at the specified key
	PutFile(bucket, key, path string) error

	// PutDirectory puts a complete directory into a bucket key prefix, with each file in the directory
	// a separate key in the bucket.
	PutDirectory(bucket, key, path string) error

	// GetFile downloads a file to a local file path
	GetFile(bucket, key, path string) error

	// OpenFile opens a file for much lower disk and memory usage that GetFile
	OpenFile(bucket, key string) (io.ReadCloser, error)

	// KeyExists checks if object exists (and if we have permission to access)
	KeyExists(bucket, key string) (bool, error)

	// Delete deletes the key from the bucket
	Delete(bucket, key string) error

	// GetDirectory downloads a directory to a local file path
	GetDirectory(bucket, key, path string) error

	// ListDirectory list the contents of a directory/bucket
	ListDirectory(bucket, keyPrefix string) ([]string, error)

	// IsDirectory tests if the key is acting like an s3 directory
	IsDirectory(bucket, key string) (bool, error)

	// BucketExists returns whether a bucket exists
	BucketExists(bucket string) (bool, error)

	// MakeBucket creates a bucket with name bucketName and options opts
	MakeBucket(bucketName string, opts minio.MakeBucketOptions) error
}

type EncryptOpts struct {
	KmsKeyId              string
	KmsEncryptionContext  string
	Enabled               bool
	ServerSideCustomerKey string
}

// AddressingStyle is a type of bucket (and also its content) addressing used by the S3 client and supported by the server
type AddressingStyle int

const (
	AutoDetectStyle AddressingStyle = iota
	PathStyle
	VirtualHostedStyle
)

type S3ClientOpts struct {
	Endpoint        string
	AddressingStyle AddressingStyle
	Region          string
	Secure          bool
	Transport       http.RoundTripper
	AccessKey       string
	SecretKey       string
	Trace           bool
	RoleARN         string
	RoleSessionName string
	UseSDKCreds     bool
	EncryptOpts     EncryptOpts
}

type s3client struct {
	S3ClientOpts
	minioClient *minio.Client
	ctx         context.Context
}

var _ S3Client = &s3client{}

// Get AWS credentials based on default order from aws SDK
func GetAWSCredentials(opts S3ClientOpts) (*credentials.Credentials, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(opts.Region)},
		SharedConfigState: session.SharedConfigEnable,
	}))

	value, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	return credentials.NewStaticV4(value.AccessKeyID, value.SecretAccessKey, value.SessionToken), nil
}

// GetAssumeRoleCredentials gets Assumed role credentials
func GetAssumeRoleCredentials(opts S3ClientOpts) (*credentials.Credentials, error) {
	sess := session.Must(session.NewSession())

	// Create the credentials from AssumeRoleProvider to assume the role
	// referenced by the "myRoleARN" ARN. Prompt for MFA token from stdin.

	creds := stscreds.NewCredentials(sess, opts.RoleARN)
	value, err := creds.Get()
	if err != nil {
		return nil, err
	}
	return credentials.NewStaticV4(value.AccessKeyID, value.SecretAccessKey, value.SessionToken), nil
}

func GetCredentials(opts S3ClientOpts) (*credentials.Credentials, error) {
	if opts.AccessKey != "" && opts.SecretKey != "" {
		log.WithField("endpoint", opts.Endpoint).Info("Creating minio client using static credentials")
		return credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""), nil
	} else if opts.RoleARN != "" {
		log.WithField("roleArn", opts.RoleARN).Info("Creating minio client using assumed-role credentials")
		return GetAssumeRoleCredentials(opts)
	} else if opts.UseSDKCreds {
		log.Info("Creating minio client using AWS SDK credentials")
		return GetAWSCredentials(opts)
	} else {
		log.Info("Creating minio client using IAM role")
		return credentials.NewIAM(nullIAMEndpoint), nil
	}
}

// GetDefaultTransport returns minio's default transport
func GetDefaultTransport(opts S3ClientOpts) (*http.Transport, error) {
	return minio.DefaultTransport(opts.Secure)
}

// NewS3Client instantiates a new S3 client object backed
func NewS3Client(ctx context.Context, opts S3ClientOpts) (S3Client, error) {
	s3cli := s3client{
		S3ClientOpts: opts,
	}
	s3cli.AccessKey = strings.TrimSpace(s3cli.AccessKey)
	s3cli.SecretKey = strings.TrimSpace(s3cli.SecretKey)
	var minioClient *minio.Client
	var err error

	credentials, err := GetCredentials(opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var bucketLookupType minio.BucketLookupType
	if s3cli.AddressingStyle == PathStyle {
		bucketLookupType = minio.BucketLookupPath
	} else if s3cli.AddressingStyle == VirtualHostedStyle {
		bucketLookupType = minio.BucketLookupDNS
	} else {
		bucketLookupType = minio.BucketLookupAuto
	}
	minioOpts := &minio.Options{Creds: credentials, Secure: s3cli.Secure, Transport: opts.Transport, Region: s3cli.Region, BucketLookup: bucketLookupType}
	minioClient, err = minio.New(s3cli.Endpoint, minioOpts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if opts.Trace {
		minioClient.TraceOn(log.StandardLogger().Out)
	}

	if opts.EncryptOpts.KmsKeyId != "" && opts.EncryptOpts.ServerSideCustomerKey != "" {
		return nil, errors.New("EncryptOpts.KmsKeyId and EncryptOpts.SSECPassword cannot be set together")
	}

	if opts.EncryptOpts.ServerSideCustomerKey != "" && !opts.Secure {
		return nil, errors.New("Secure must be set if EncryptOpts.SSECPassword is set")
	}

	s3cli.ctx = ctx
	s3cli.minioClient = minioClient

	return &s3cli, nil
}

// PutFile puts a single file to a bucket at the specified key
func (s *s3client) PutFile(bucket, key, path string) error {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": key, "path": path}).Info("Saving file to s3")
	// NOTE: minio will detect proper mime-type based on file extension

	encOpts, err := s.EncryptOpts.buildServerSideEnc(bucket, key)

	if err != nil {
		return errors.WithStack(err)
	}

	_, err = s.minioClient.FPutObject(s.ctx, bucket, key, path, minio.PutObjectOptions{ServerSideEncryption: encOpts})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *s3client) BucketExists(bucketName string) (bool, error) {
	log.WithField("bucket", bucketName).Info("Checking if bucket exists")
	result, err := s.minioClient.BucketExists(s.ctx, bucketName)
	return result, errors.WithStack(err)
}

func (s *s3client) MakeBucket(bucketName string, opts minio.MakeBucketOptions) error {
	log.WithFields(log.Fields{"bucket": bucketName, "region": opts.Region, "objectLocking": opts.ObjectLocking}).Info("Creating bucket")
	err := s.minioClient.MakeBucket(s.ctx, bucketName, opts)

	if err != nil {
		return errors.WithStack(err)
	}

	err = s.setBucketEnc(bucketName)
	return errors.WithStack(err)
}

type uploadTask struct {
	key  string
	path string
}

func generatePutTasks(keyPrefix, rootPath string) chan uploadTask {
	rootPath = filepath.Clean(rootPath) + string(os.PathSeparator)
	uploadTasks := make(chan uploadTask)
	go func() {
		_ = filepath.Walk(rootPath, func(localPath string, fi os.FileInfo, _ error) error {
			relPath := strings.TrimPrefix(localPath, rootPath)
			if fi.IsDir() {
				return nil
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			t := uploadTask{
				key:  path.Join(keyPrefix, relPath),
				path: localPath,
			}
			uploadTasks <- t
			return nil
		})
		close(uploadTasks)
	}()
	return uploadTasks
}

// PutDirectory puts a complete directory into a bucket key prefix, with each file in the directory
// a separate key in the bucket.
func (s *s3client) PutDirectory(bucket, key, path string) error {
	for putTask := range generatePutTasks(key, path) {
		err := s.PutFile(bucket, putTask.key, putTask.path)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetFile downloads a file to a local file path
func (s *s3client) GetFile(bucket, key, path string) error {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": key, "path": path}).Info("Getting file from s3")

	encOpts, err := s.EncryptOpts.buildServerSideEnc(bucket, key)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.minioClient.FGetObject(s.ctx, bucket, key, path, minio.GetObjectOptions{ServerSideEncryption: encOpts})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// OpenFile opens a file for reading
func (s *s3client) OpenFile(bucket, key string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": key}).Info("Opening file from s3")

	encOpts, err := s.EncryptOpts.buildServerSideEnc(bucket, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	f, err := s.minioClient.GetObject(s.ctx, bucket, key, minio.GetObjectOptions{ServerSideEncryption: encOpts})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// the call above doesn't return an error in the case that the key doesn't exist, but by calling Stat() it will
	_, err = f.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return f, nil
}

// checks if object exists (and if we have permission to access)
func (s *s3client) KeyExists(bucket, key string) (bool, error) {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": key}).Info("Checking key exists from s3")

	encOpts, err := s.EncryptOpts.buildServerSideEnc(bucket, key)
	if err != nil {
		return false, errors.WithStack(err)
	}

	_, err = s.minioClient.StatObject(s.ctx, bucket, key, minio.StatObjectOptions{ServerSideEncryption: encOpts})
	if err == nil {
		return true, nil
	}
	if IsS3ErrCode(err, "NoSuchKey") {
		return false, nil
	}

	return false, errors.WithStack(err)
}

func (s *s3client) Delete(bucket, key string) error {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": key}).Info("Deleting object from s3")
	return s.minioClient.RemoveObject(s.ctx, bucket, key, minio.RemoveObjectOptions{})
}

// GetDirectory downloads a s3 directory to a local path
func (s *s3client) GetDirectory(bucket, keyPrefix, path string) error {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": keyPrefix, "path": path}).Info("Getting directory from s3")
	keys, err := s.ListDirectory(bucket, keyPrefix)
	if err != nil {
		return err
	}

	for _, objKey := range keys {
		relKeyPath := strings.TrimPrefix(objKey, keyPrefix)
		localPath := filepath.Join(path, relKeyPath)

		encOpts, err := s.EncryptOpts.buildServerSideEnc(bucket, objKey)
		if err != nil {
			return errors.WithStack(err)
		}

		err = s.minioClient.FGetObject(s.ctx, bucket, objKey, localPath, minio.GetObjectOptions{ServerSideEncryption: encOpts})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// IsDirectory tests if the key is acting like a s3 directory. This just means it has at least one
// object which is prefixed with the given key
func (s *s3client) IsDirectory(bucket, keyPrefix string) (bool, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	if keyPrefix != "" {
		keyPrefix = filepath.Clean(keyPrefix) + "/"
		if os.PathSeparator == '\\' {
			keyPrefix = strings.ReplaceAll(keyPrefix, "\\", "/")
		}
	}

	listOpts := minio.ListObjectsOptions{
		Prefix:    keyPrefix,
		Recursive: false,
	}
	objCh := s.minioClient.ListObjects(s.ctx, bucket, listOpts)
	for obj := range objCh {
		if obj.Err != nil {
			return false, errors.WithStack(obj.Err)
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (s *s3client) ListDirectory(bucket, keyPrefix string) ([]string, error) {
	log.WithFields(log.Fields{"endpoint": s.Endpoint, "bucket": bucket, "key": keyPrefix}).Info("Listing directory from s3")

	if keyPrefix != "" {
		keyPrefix = filepath.Clean(keyPrefix) + "/"
		if os.PathSeparator == '\\' {
			keyPrefix = strings.ReplaceAll(keyPrefix, "\\", "/")
		}
	}

	doneCh := make(chan struct{})
	defer close(doneCh)
	listOpts := minio.ListObjectsOptions{
		Prefix:    keyPrefix,
		Recursive: true,
	}
	var out []string
	objCh := s.minioClient.ListObjects(s.ctx, bucket, listOpts)
	for obj := range objCh {
		if obj.Err != nil {
			return nil, errors.WithStack(obj.Err)
		}
		if strings.HasSuffix(obj.Key, "/") {
			// When a dir is created through AWS S3 console, a nameless obj will be created
			// automatically, its key will be {dir_name} + "/". This obj does not display in the
			// console, but you can see it when using aws cli.
			// If obj.Key ends with "/" means it's a dir obj, we need to skip it, otherwise it
			// will be downloaded as a regular file with the same name as the dir, and it will
			// creates error when downloading the files under the dir.
			continue
		}
		out = append(out, obj.Key)
	}
	return out, nil
}

// IsS3ErrCode returns if the supplied error is of a specific S3 error code
func IsS3ErrCode(err error, code string) bool {
	err = errors.Cause(err)
	if minioErr, ok := err.(minio.ErrorResponse); ok {
		return minioErr.Code == code
	}
	return false
}

// setBucketEnc sets the encryption options on a bucket
func (s *s3client) setBucketEnc(bucketName string) error {
	if !s.EncryptOpts.Enabled {
		return nil
	}

	var config *sse.Configuration
	if s.EncryptOpts.KmsKeyId != "" {
		config = sse.NewConfigurationSSEKMS(s.EncryptOpts.KmsKeyId)
	} else {
		config = sse.NewConfigurationSSES3()
	}

	log.WithFields(log.Fields{"KmsKeyId": s.EncryptOpts.KmsKeyId, "bucketName": bucketName}).Info("Setting Bucket Encryption")
	err := s.minioClient.SetBucketEncryption(s.ctx, bucketName, config)
	return err
}

// buildServerSideEnc creates the minio encryption options when putting encrypted items in a bucket
func (e *EncryptOpts) buildServerSideEnc(bucket, key string) (encrypt.ServerSide, error) {
	if e == nil || !e.Enabled {
		return nil, nil
	}

	if e.ServerSideCustomerKey != "" {
		encryption := encrypt.DefaultPBKDF([]byte(e.ServerSideCustomerKey), []byte(bucket+key))

		return encryption, nil
	}

	if e.KmsKeyId != "" {
		encryptionCtx, err := parseKMSEncCntx(e.KmsEncryptionContext)

		if err != nil {
			return nil, errors.Wrap(err, "failed to parse KMS encryption context")
		}

		if encryptionCtx == nil {
			// To overcome a limitation in Minio which checks interface{} == nil.
			kms, err := encrypt.NewSSEKMS(e.KmsKeyId, nil)

			if err != nil {
				return nil, err
			}

			return kms, nil
		}

		kms, err := encrypt.NewSSEKMS(e.KmsKeyId, encryptionCtx)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		return kms, nil
	}

	return encrypt.NewSSE(), nil
}

// parseKMSEncCntx validates if kmsEncCntx is a valid JSON
func parseKMSEncCntx(kmsEncCntx string) (*string, error) {
	if kmsEncCntx == "" {
		return nil, nil
	}

	jsonKMSEncryptionContext, err := json.Marshal(json.RawMessage(kmsEncCntx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal KMS encryption context")
	}

	parsedKMSEncryptionContext := base64.StdEncoding.EncodeToString([]byte(jsonKMSEncryptionContext))

	return &parsedKMSEncryptionContext, nil
}
