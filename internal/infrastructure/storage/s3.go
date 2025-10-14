package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	bucketName    string
	publicBaseURL string
	s3Client      *s3.Client
}

type S3Opts struct {
	AccessKey, SecretKey, Region, Endpoint string
	Bucket                                 string
	UsePathStyle                           bool
	PublicBase                             string
}

func NewS3(ctx context.Context, opts S3Opts) (*S3Client, error) {
	var configLoaders []func(*awscfg.LoadOptions) error

	if opts.AccessKey != "" && opts.SecretKey != "" {
		configLoaders = append(configLoaders, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, ""),
		))
	}

	if opts.Region != "" {
		configLoaders = append(configLoaders, awscfg.WithRegion(opts.Region))
	}

	awsConfig, err := awscfg.LoadDefaultConfig(ctx, configLoaders...)
	if err != nil {
		return nil, err
	}

	clientOptions := []func(*s3.Options){}

	if opts.Endpoint != "" {
		endpointURL, _ := url.Parse(opts.Endpoint)
		clientOptions = append(clientOptions, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(endpointURL.String())
			options.Region = opts.Region
			options.UsePathStyle = opts.UsePathStyle
		})
	} else if opts.Region != "" {
		clientOptions = append(clientOptions, func(options *s3.Options) {
			options.Region = opts.Region
		})
	}

	s3APIClient := s3.NewFromConfig(awsConfig, clientOptions...)

	return &S3Client{
		bucketName:    opts.Bucket,
		publicBaseURL: opts.PublicBase,
		s3Client:      s3APIClient,
	}, nil
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

func ObjectKeyFromHash(hashHex, filename string) string {
	extension := ""
	if dotIndex := strings.LastIndex(filename, "."); dotIndex >= 0 {
		extension = strings.ToLower(filename[dotIndex:])
	}

	prefix := "00"
	if len(hashHex) >= 2 {
		prefix = hashHex[:2]
	}

	return fmt.Sprintf("%s/%s%s", prefix, hashHex, extension)
}
