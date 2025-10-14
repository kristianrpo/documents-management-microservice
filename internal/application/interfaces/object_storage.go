package interfaces

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	Put(ctx context.Context, body io.Reader, objectKey, contentType string) error
	PublicURL(objectKey string) string
	Bucket() string
	Delete(ctx context.Context, objectKey string) error
	// GeneratePresignedURL generates a temporary pre-signed URL for secure access to an object
	GeneratePresignedURL(ctx context.Context, objectKey string, expiration time.Duration) (string, error)
}
