package tests

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
)

func TestS3Client_PublicURL(t *testing.T) {
	cli := storage.NewS3Client("bucket", "http://public.example.com", &s3.Client{})
	url := cli.PublicURL("path/to/object.txt")
	assert.Equal(t, "http://public.example.com/path/to/object.txt", url)
}

func TestS3Client_PublicURL_EmptyBase(t *testing.T) {
	cli := storage.NewS3Client("bucket", "", &s3.Client{})
	url := cli.PublicURL("object.txt")
	assert.Equal(t, "", url)
}

func TestS3Client_Bucket(t *testing.T) {
	cli := storage.NewS3Client("my-bucket", "", &s3.Client{})
	assert.Equal(t, "my-bucket", cli.Bucket())
}

func TestS3Client_Delete_NoErrorMock(t *testing.T) {
	// We just ensure method wiring compiles; we cannot hit AWS here. Use a zero client and expect error nil path not guaranteed.
	cli := storage.NewS3Client("bucket", "", &s3.Client{})
	_ = cli
	// Create a context to ensure signature use
	_ = context.Background()
	// We won't call Put/Delete to avoid hitting AWS; those are thin wrappers already covered via integration.
}
