package storage

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"sync"
)

// S3Client implements the ObjectStorage interface using AWS S3 or compatible storage (MinIO)
type S3Client struct {
	bucketName    string
	publicBaseURL string
	s3Client      *s3.Client
	ensureOnce    sync.Once
}

// NewS3Client creates a new S3 client for object storage operations
func NewS3Client(bucketName, publicBaseURL string, s3APIClient *s3.Client) *S3Client {
	return &S3Client{
		bucketName:    bucketName,
		publicBaseURL: publicBaseURL,
		s3Client:      s3APIClient,
	}
}

// Put uploads an object to S3 with the specified key and content type
func (client *S3Client) Put(ctx context.Context, body io.Reader, objectKey, contentType string) error {
	// Ensure bucket exists once per process (best-effort)
	client.ensureOnce.Do(func() {
		_ = client.ensureBucket(ctx)
	})

	// First attempt
	_, err := client.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(client.bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err == nil {
		return nil
	}

	// If the error is NoSuchBucket/NotFound, try to create the bucket and retry once
	if isNoSuchBucket(err) {
		if cerr := client.ensureBucket(ctx); cerr == nil {
			// Rewind if possible before retrying
			if seeker, ok := body.(interface {
				Seek(int64, int) (int64, error)
			}); ok {
				_, _ = seeker.Seek(0, 0)
			}
			_, err2 := client.s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(client.bucketName),
				Key:         aws.String(objectKey),
				Body:        body,
				ContentType: aws.String(contentType),
				ACL:         types.ObjectCannedACLPrivate,
			})
			if err2 == nil {
				return nil
			}
			return err2
		}
	}
	return err
}

// PublicURL constructs the public URL for accessing an object
func (client *S3Client) PublicURL(objectKey string) string {
	if client.publicBaseURL == "" {
		return ""
	}
	return strings.TrimRight(client.publicBaseURL, "/") + "/" + path.Clean(objectKey)
}

// Bucket returns the name of the S3 bucket
func (client *S3Client) Bucket() string {
	return client.bucketName
}

// Delete removes an object from S3
func (client *S3Client) Delete(ctx context.Context, objectKey string) error {
	_, err := client.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(client.bucketName),
		Key:    aws.String(objectKey),
	})
	return err
}

// GeneratePresignedURL creates a temporary pre-signed URL for secure access to an object
func (client *S3Client) GeneratePresignedURL(ctx context.Context, objectKey string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(client.bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", err
	}

	return request.URL, nil
}

// ensureBucket checks if the bucket exists and creates it if missing. Best-effort: returns nil
// if the bucket already exists or is created successfully. Returns error only on definitive failures.
func (client *S3Client) ensureBucket(ctx context.Context) error {
	// HeadBucket returns 404/NotFound if bucket does not exist
	_, err := client.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(client.bucketName)})
	if err == nil {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		if code == "NoSuchBucket" || code == "NotFound" {
			// Try to create the bucket
			_, cErr := client.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(client.bucketName),
			})
			if cErr == nil {
				return nil
			}
			if errors.As(cErr, &apiErr) {
				switch apiErr.ErrorCode() {
				case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
					return nil
				}
			}
			return cErr
		}
	}
	// For non-API errors or other codes (e.g., auth), do not block normal flow here
	return nil
}

// isNoSuchBucket returns true if the error corresponds to a missing bucket
func isNoSuchBucket(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NoSuchBucket" || code == "NotFound"
	}
	return false
}
