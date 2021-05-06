package s3

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	// GetDirectory downloads a directory to a local file path
	GetDirectory(bucket, key, path string) error

	// ListDirectory list the contents of a directory/bucket
	ListDirectory(bucket, keyPrefix string) ([]string, error)

	// IsDirectory tests if the key is acting like a s3 directory
	IsDirectory(bucket, key string) (bool, error)

	// BucketExists returns whether a bucket exists
	BucketExists(bucket string) (bool, error)

	// MakeBucket creates a bucket with name bucketName and options opts
	MakeBucket(bucketName string, opts minio.MakeBucketOptions) error
}

type S3ClientOpts struct {
	Endpoint        string
	Region          string
	Secure          bool
	AccessKey       string
	SecretKey       string
	Trace           bool
	RoleARN         string
	RoleSessionName string
	UseSDKCreds     bool
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
		log.Infof("Creating minio client %s using static credentials", opts.Endpoint)
		return credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""), nil
	} else if opts.RoleARN != "" {
		log.Infof("Creating minio client %s using assumed-role credentials", opts.RoleARN)
		return GetAssumeRoleCredentials(opts)
	} else if opts.UseSDKCreds {
		log.Infof("Creating minio client using AWS SDK credentials")
		return GetAWSCredentials(opts)
	} else {
		log.Infof("Creating minio client using IAM role")
		return credentials.NewIAM(nullIAMEndpoint), nil
	}
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

	minioOpts := &minio.Options{Creds: credentials, Secure: s3cli.Secure, Region: s3cli.Region}
	minioClient, err = minio.New(s3cli.Endpoint, minioOpts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if opts.Trace {
		minioClient.TraceOn(log.StandardLogger().Out)
	}
	s3cli.ctx = ctx
	s3cli.minioClient = minioClient
	return &s3cli, nil
}

// PutFile puts a single file to a bucket at the specified key
func (s *s3client) PutFile(bucket, key, path string) error {
	log.Infof("Saving from %s to s3 (endpoint: %s, bucket: %s, key: %s)", path, s.Endpoint, bucket, key)
	// NOTE: minio will detect proper mime-type based on file extension
	_, err := s.minioClient.FPutObject(s.ctx, bucket, key, path, minio.PutObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *s3client) BucketExists(bucketName string) (bool, error) {
	log.Infof("Checking if bucket %s exists.", bucketName)
	result, err := s.minioClient.BucketExists(s.ctx, bucketName)
	return result, errors.WithStack(err)
}

func (s *s3client) MakeBucket(bucketName string, opts minio.MakeBucketOptions) error {
	log.Infof("Creating bucket: %s. (Region: %s, ObjectLocking: %t)", bucketName, opts.Region, opts.ObjectLocking)
	err := s.minioClient.MakeBucket(s.ctx, bucketName, opts)
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
	log.Infof("Getting from s3 (endpoint: %s, bucket: %s, key: %s) to %s", s.Endpoint, bucket, key, path)
	err := s.minioClient.FGetObject(s.ctx, bucket, key, path, minio.GetObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetDirectory downloads a s3 directory to a local path
func (s *s3client) GetDirectory(bucket, keyPrefix, path string) error {
	log.Infof("Getting directory from s3 (endpoint: %s, bucket: %s, key: %s) to %s", s.Endpoint, bucket, keyPrefix, path)

	keys, err := s.ListDirectory(bucket, keyPrefix)
	if err != nil {
		return err
	}

	for _, objKey := range keys {
		relKeyPath := strings.TrimPrefix(objKey, keyPrefix)
		localPath := filepath.Join(path, relKeyPath)
		err := s.minioClient.FGetObject(s.ctx, bucket, objKey, localPath, minio.GetObjectOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// IsDirectory tests if the key is acting like a s3 directory. This just means it has at least one
// object which is prefixed with the given key
func (s *s3client) IsDirectory(bucket, key string) (bool, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	listOpts := minio.ListObjectsOptions{
		Prefix:    key,
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
	log.Infof("Listing directory from s3 (endpoint: %s, bucket: %s, key: %s)", s.Endpoint, bucket, keyPrefix)

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
