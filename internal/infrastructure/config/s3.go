package config

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/storage"
)

type S3Opts struct {
	AccessKey, SecretKey, Region, Endpoint string
	Bucket                                 string
	UsePathStyle                           bool
	PublicBase                             string
}

func NewS3(ctx context.Context, opts S3Opts) (*storage.S3Client, error) {
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

	return storage.NewS3Client(opts.Bucket, opts.PublicBase, s3APIClient), nil
}

func NewS3Client(ctx context.Context, cfg Config) (*storage.S3Client, error) {
	opts := S3Opts{
		AccessKey:    cfg.AWSAccessKey,
		SecretKey:    cfg.AWSSecretKey,
		Region:       cfg.AWSRegion,
		Endpoint:     cfg.S3Endpoint,
		Bucket:       cfg.S3Bucket,
		UsePathStyle: cfg.S3UsePath,
		PublicBase:   cfg.S3PublicBase,
	}
	return NewS3(ctx, opts)
}
