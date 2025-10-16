package storage

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client implements the ObjectStorage interface using AWS S3 or compatible storage (MinIO)
type S3Client struct {
	bucketName    string
	publicBaseURL string
	s3Client      *s3.Client
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
	_, err := client.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(client.bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
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
