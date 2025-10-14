package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port string

	DynamoDBTable    string
	DynamoDBEndpoint string

	AWSAccessKey string
	AWSSecretKey string
	AWSRegion    string
	S3Bucket     string
	S3Endpoint   string
	S3UsePath    bool
	S3PublicBase string

	ReadHeaderTimeout time.Duration
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
func getbool(k string) bool { return os.Getenv(k) == "true" }

func Load() *Config {
	port := ":" + getenv("APP_PORT", "8080")

	return &Config{
		Port:             port,
		DynamoDBTable:    getenv("DYNAMODB_TABLE", "documents"),
		DynamoDBEndpoint: getenv("DYNAMODB_ENDPOINT", ""),
		AWSAccessKey:     getenv("AWS_ACCESS_KEY_ID", "local"),
		AWSSecretKey:     getenv("AWS_SECRET_ACCESS_KEY", "local"),
		AWSRegion:        getenv("AWS_REGION", "us-east-1"),
		S3Bucket:         getenv("S3_BUCKET", "documents"),
		S3Endpoint:       getenv("S3_ENDPOINT", ""),
		S3UsePath:        getbool("S3_USE_PATH_STYLE"),
		S3PublicBase:     getenv("S3_PUBLIC_BASE_URL", ""),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (c *Config) Validate() error {
	if c.S3Bucket == "" { return errors.New("S3_BUCKET required") }
	if c.S3Endpoint != "" && !c.S3UsePath { return errors.New("S3_USE_PATH_STYLE=true required when using S3_ENDPOINT (MinIO)") }
	if c.S3Endpoint == "" && c.AWSRegion == "" { return errors.New("AWS_REGION required for AWS S3") }
	return nil
}
