package storage

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	bucketName    string
	publicBaseURL string
	s3Client      *s3.Client
}

func NewS3Client(bucketName, publicBaseURL string, s3APIClient *s3.Client) *S3Client {
	return &S3Client{
		bucketName:    bucketName,
		publicBaseURL: publicBaseURL,
		s3Client:      s3APIClient,
	}
}

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

func (client *S3Client) PublicURL(objectKey string) string {
	if client.publicBaseURL == "" {
		return ""
	}
	return strings.TrimRight(client.publicBaseURL, "/") + "/" + path.Clean(objectKey)
}

func (client *S3Client) Bucket() string {
	return client.bucketName
}

func (client *S3Client) Delete(ctx context.Context, objectKey string) error {
	_, err := client.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(client.bucketName),
		Key:    aws.String(objectKey),
	})
	return err
}
