package s3

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewS3Client tests the s3 construtor
func TestNewS3Client(t *testing.T) {
	opts := S3ClientOpts{
		Endpoint:        "foo.com",
		Region:          "us-south-3",
		Secure:          false,
		AccessKey:       "key",
		SecretKey:       "secret",
		Trace:           true,
		RoleARN:         "",
		RoleSessionName: "",
		UseSDKCreds:     false,
		EncryptOpts:     EncryptOpts{Enabled: true, ServerSideCustomerKey: "", KmsKeyId: "", KmsEncryptionContext: ""},
	}
	s3If, err := NewS3Client(context.Background(), opts)
	assert.NoError(t, err)
	s3cli := s3If.(*s3client)
	assert.Equal(t, opts.Endpoint, s3cli.Endpoint)
	assert.Equal(t, opts.Region, s3cli.Region)
	assert.Equal(t, opts.Secure, s3cli.Secure)
	assert.Equal(t, opts.AccessKey, s3cli.AccessKey)
	assert.Equal(t, opts.Trace, s3cli.Trace)
	assert.Equal(t, opts.EncryptOpts, s3cli.EncryptOpts)
	// s3cli.minioClient.
	// 	s3client.minioClient
}

// TestNewS3Client tests the s3 construtor
func TestNewS3ClientWithDiff(t *testing.T) {
	t.Run("IAMRole", func(t *testing.T) {
		opts := S3ClientOpts{
			Endpoint: "foo.com",
			Region:   "us-south-3",
			Secure:   false,
			Trace:    true,
		}
		s3If, err := NewS3Client(context.Background(), opts)
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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

func TestPutDirectorySuccess(t *testing.T) {
	var requests []http.Request
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, *r)
		fmt.Fprintf(w, "OK")
	}))
	defer svr.Close()
	opts := S3ClientOpts{
		Endpoint:  strings.Split(svr.URL, "http://")[1],
		Region:    "us-west-2",
		AccessKey: "ABCDEF",
		SecretKey: "ABCDEF",
	}
	s3, err := NewS3Client(context.Background(), opts)
	assert.NoError(t, err)
	testDirectory, err := ioutil.TempDir("/tmp", "argo-s3-test-*")
	assert.NoError(t, err)
	filesCount := 50
	for i := 0; i < filesCount; i++ {
		_, err := ioutil.TempFile(testDirectory, "*")
		assert.NoError(t, err)
	}

	err = s3.PutDirectory("test-bucket", "test-key", testDirectory)
	assert.NoError(t, err)
	svr.Close()

	assert.Len(t, requests, filesCount, "Expected %d requests but got %d", filesCount, len(requests))
}

func TestPutDirectoryErrorStopsWorkers(t *testing.T) {
	var requests []http.Request
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, *r)
		w.WriteHeader(500)
		fmt.Fprintf(w, "Server Error")
	}))
	defer svr.Close()
	opts := S3ClientOpts{
		Endpoint:  strings.Split(svr.URL, "http://")[1],
		Region:    "us-west-2",
		AccessKey: "ABCDEF",
		SecretKey: "ABCDEF",
	}
	s3, err := NewS3Client(context.Background(), opts)
	assert.NoError(t, err)
	testDirectory, err := ioutil.TempDir("/tmp", "argo-s3-test-*")
	assert.NoError(t, err)
	for i := 0; i < 50; i++ {
		_, err := ioutil.TempFile(testDirectory, "*")
		assert.NoError(t, err)
	}

	err = s3.PutDirectory("test-bucket", "test-key", testDirectory)
	assert.ErrorContains(t, err, "Server Error")
	var uniqueURLRequests []http.Request
	for _, request := range requests {
		isUnique := true
		for _, uniqueRequest := range uniqueURLRequests {
			if uniqueRequest.RequestURI == request.RequestURI {
				isUnique = false
				break
			}
		}
		if isUnique {
			uniqueURLRequests = append(uniqueURLRequests, request)
		}
	}

	assert.Len(t, uniqueURLRequests, 10, "Expected %d requests but got %d", 10, len(uniqueURLRequests))
}
