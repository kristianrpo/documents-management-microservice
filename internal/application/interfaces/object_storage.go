package interfaces

import (
	"context"
	"io"
	"time"
)

// ObjectStorage defines the interface for object storage operations (S3, MinIO, etc.)
type ObjectStorage interface {
	// Put uploads an object to storage with the specified key and content type
	Put(ctx context.Context, body io.Reader, objectKey, contentType string) error
	
	// PublicURL returns the public URL for accessing an object
	PublicURL(objectKey string) string
	
	// Bucket returns the name of the storage bucket
	Bucket() string
	
	// Delete removes an object from storage
	Delete(ctx context.Context, objectKey string) error
	
	// GeneratePresignedURL generates a temporary pre-signed URL for secure access to an object
	GeneratePresignedURL(ctx context.Context, objectKey string, expiration time.Duration) (string, error)
}
