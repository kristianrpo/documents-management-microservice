package config

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/storage"
)

// NewS3Client creates a new S3 client with the provided configuration
// Supports both AWS S3 and S3-compatible storage (MinIO)
func NewS3Client(ctx context.Context, cfg Config) (*storage.S3Client, error) {
	var configLoaders []func(*awscfg.LoadOptions) error

	if cfg.AWSAccessKey != "" && cfg.AWSSecretKey != "" {
		configLoaders = append(configLoaders, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWSAccessKey, cfg.AWSSecretKey, ""),
		))
	}

	if cfg.AWSRegion != "" {
		configLoaders = append(configLoaders, awscfg.WithRegion(cfg.AWSRegion))
	}

	awsConfig, err := awscfg.LoadDefaultConfig(ctx, configLoaders...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOptions := []func(*s3.Options){}

	if cfg.S3Endpoint != "" {
		endpointURL, err := url.Parse(cfg.S3Endpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid S3 endpoint URL: %w", err)
		}
		clientOptions = append(clientOptions, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(endpointURL.String())
			options.Region = cfg.AWSRegion
			options.UsePathStyle = cfg.S3UsePath
		})
	} else if cfg.AWSRegion != "" {
		clientOptions = append(clientOptions, func(options *s3.Options) {
			options.Region = cfg.AWSRegion
		})
	}

	s3APIClient := s3.NewFromConfig(awsConfig, clientOptions...)

	return storage.NewS3Client(cfg.S3Bucket, cfg.S3PublicBase, s3APIClient), nil
}
