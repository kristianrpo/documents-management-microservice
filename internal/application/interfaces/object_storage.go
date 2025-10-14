package interfaces

import (
    "context"
    "io"
)

type ObjectStorage interface {
    Put(ctx context.Context, body io.Reader, objectKey, contentType string) error
    PublicURL(objectKey string) string
    Bucket() string
}
