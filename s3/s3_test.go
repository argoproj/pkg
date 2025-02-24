package s3

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewS3Client tests the s3 constructor
func TestNewS3Client(t *testing.T) {
	opts := S3ClientOpts{
		Endpoint:        "foo.com",
		Region:          "us-south-3",
		Secure:          false,
		Transport:       http.DefaultTransport,
		AccessKey:       "key",
		SecretKey:       "secret",
		SessionToken:    "",
		Trace:           true,
		RoleARN:         "",
		RoleSessionName: "",
		UseSDKCreds:     false,
		EncryptOpts:     EncryptOpts{Enabled: true, ServerSideCustomerKey: "", KmsKeyId: "", KmsEncryptionContext: ""},
	}
	s3If, err := NewS3Client(context.Background(), opts)
	require.NoError(t, err)
	s3cli := s3If.(*s3client)
	assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
	assert.Equal(t, opts.Region, s3cli.Region)
	assert.Equal(t, opts.Secure, s3cli.Secure)
	assert.Equal(t, opts.Transport, s3cli.Transport)
	assert.Equal(t, opts.AccessKey, s3cli.AccessKey)
	assert.Equal(t, opts.SessionToken, s3cli.SessionToken)
	assert.Equal(t, opts.Trace, s3cli.Trace)
	assert.Equal(t, opts.EncryptOpts, s3cli.EncryptOpts)
	assert.Equal(t, opts.AddressingStyle, s3cli.AddressingStyle)
	// s3cli.minioClient.
	// 	s3client.minioClient
}

// TestNewS3Client tests the S3 constructor using ephemeral credentials
func TestNewS3ClientEphemeral(t *testing.T) {
	opts := S3ClientOpts{
		Endpoint:     "foo.com",
		Region:       "us-south-3",
		AccessKey:    "key",
		SecretKey:    "secret",
		SessionToken: "sessionToken",
	}
	s3If, err := NewS3Client(context.Background(), opts)
	require.NoError(t, err)
	s3cli := s3If.(*s3client)
	assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
	assert.Equal(t, opts.Region, s3cli.Region)
	assert.Equal(t, opts.AccessKey, s3cli.AccessKey)
	assert.Equal(t, opts.SecretKey, s3cli.SecretKey)
	assert.Equal(t, opts.SessionToken, s3cli.SessionToken)
}

// TestNewS3Client tests the s3 constructor
func TestNewS3ClientWithDiff(t *testing.T) {
	t.Run("IAMRole", func(t *testing.T) {
		opts := S3ClientOpts{
			Endpoint: "foo.com",
			Region:   "us-south-3",
			Secure:   false,
			Trace:    true,
		}
		s3If, err := NewS3Client(context.Background(), opts)
		require.NoError(t, err)
		s3cli := s3If.(*s3client)
		assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
		assert.Equal(t, opts.Region, s3cli.Region)
		assert.Equal(t, opts.Trace, s3cli.Trace)
		assert.Equal(t, opts.Endpoint, s3cli.minioClient.EndpointURL().Host)
	})
	t.Run("AssumeIAMRole", func(t *testing.T) {
		t.SkipNow()
		opts := S3ClientOpts{
			Endpoint: "foo.com",
			Region:   "us-south-3",
			Secure:   false,
			Trace:    true,
			RoleARN:  "01234567890123456789",
		}
		s3If, err := NewS3Client(context.Background(), opts)
		require.NoError(t, err)
		s3cli := s3If.(*s3client)
		assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
		assert.Equal(t, opts.Region, s3cli.Region)
		assert.Equal(t, opts.Trace, s3cli.Trace)
		assert.Equal(t, opts.Endpoint, s3cli.minioClient.EndpointURL().Host)
	})
}

func TestDisallowedComboOptions(t *testing.T) {
	t.Run("KMS and SSEC", func(t *testing.T) {
		opts := S3ClientOpts{
			Endpoint:    "foo.com",
			Region:      "us-south-3",
			Secure:      true,
			Trace:       true,
			EncryptOpts: EncryptOpts{Enabled: true, ServerSideCustomerKey: "PASSWORD", KmsKeyId: "00000000-0000-0000-0000-000000000000", KmsEncryptionContext: ""},
		}
		_, err := NewS3Client(context.Background(), opts)
		assert.Error(t, err)
	})

	t.Run("SSEC and InSecure", func(t *testing.T) {
		opts := S3ClientOpts{
			Endpoint:    "foo.com",
			Region:      "us-south-3",
			Secure:      false,
			Trace:       true,
			EncryptOpts: EncryptOpts{Enabled: true, ServerSideCustomerKey: "PASSWORD", KmsKeyId: "", KmsEncryptionContext: ""},
		}
		_, err := NewS3Client(context.Background(), opts)
		assert.Error(t, err)
	})
}
