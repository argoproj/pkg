package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewS3Client tests the s3 construtor
func TestNewS3Client(t *testing.T) {
	opts := S3ClientOpts{
		Endpoint:  "foo.com",
		Region:    "us-south-3",
		Secure:    false,
		AccessKey: "key",
		SecretKey: "secret",
		Trace:     true,
	}
	s3If, err := NewS3Client(opts)
	assert.NoError(t, err)
	s3cli := s3If.(*s3client)
	assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
	assert.Equal(t, opts.Region, s3cli.Region)
	assert.Equal(t, opts.Secure, s3cli.Secure)
	assert.Equal(t, opts.AccessKey, s3cli.AccessKey)
	assert.Equal(t, opts.Trace, s3cli.Trace)
	// s3cli.minioClient.
	// 	s3client.minioClient
}
