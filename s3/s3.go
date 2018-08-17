package s3

import (
	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
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
}

type S3ClientOpts struct {
	Endpoint  string
	Region    string
	Secure    bool
	AccessKey string
	SecretKey string
}

type s3client struct {
	S3ClientOpts
	minioClient *minio.Client
}

// NewS3Client instantiates a new S3 client object backed
func NewS3Client(opts S3ClientOpts) (S3Client, error) {
	s3cli := s3client{
		S3ClientOpts: opts,
	}
	var minioClient *minio.Client
	var err error
	if s3cli.AccessKey != "" {
		log.Infof("Creating minio client %s using static credentials", s3cli.Endpoint)
		if s3cli.Region != "" {
			minioClient, err = minio.NewWithRegion(
				s3cli.Endpoint, s3cli.AccessKey, s3cli.SecretKey, s3cli.Secure, s3cli.Region)
		} else {
			minioClient, err = minio.New(s3cli.Endpoint, s3cli.AccessKey, s3cli.SecretKey, s3cli.Secure)
		}
	} else {
		log.Infof("Creating minio client %s using IAM role", s3cli.Endpoint)
		credentials := credentials.NewIAM(nullIAMEndpoint)
		minioClient, err = minio.NewWithCredentials(s3cli.Endpoint, credentials, s3cli.Secure, s3cli.Region)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s3cli.minioClient = minioClient
	return &s3cli, nil
}

// PutFile puts a single file to a bucket at the specified key
func (s *s3client) PutFile(bucket, key, path string) error {
	log.Infof("Saving from %s to s3 (endpoint: %s, bucket: %s, key: %s)", path, s.Endpoint, bucket, key)
	// NOTE: minio will detect proper mime-type based on file extension
	_, err := s.minioClient.FPutObject(bucket, key, path, minio.PutObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil

}

// PutDirectory puts a complete directory into a bucket key prefix, with each file in the directory
// a separate key in the bucket.
func (s *s3client) PutDirectory(bucket, key, path string) error {
	return errors.New("not yet implemented")
}

// GetFile downloads a file to a local file path
func (s *s3client) GetFile(bucket, key, path string) error {
	log.Infof("Getting from s3 (endpoint: %s, bucket: %s, key: %s) to %s", s.Endpoint, bucket, key, path)
	err := s.minioClient.FGetObject(bucket, key, path, minio.GetObjectOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
