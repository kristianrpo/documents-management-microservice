package config_test

import (
	"os"
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("load with defaults", func(t *testing.T) {
		// Clear environment variables
		os.Clearenv()

		cfg := config.Load()

		assert.NotNil(t, cfg)
		assert.Equal(t, ":8080", cfg.Port)
		assert.Equal(t, "documents", cfg.DynamoDBTable)
		assert.Equal(t, "local", cfg.AWSAccessKey)
		assert.Equal(t, "local", cfg.AWSSecretKey)
		assert.Equal(t, "us-east-1", cfg.AWSRegion)
		assert.Equal(t, "documents", cfg.S3Bucket)
		assert.Equal(t, "", cfg.S3Endpoint)
		assert.False(t, cfg.S3UsePath)
		assert.Equal(t, "", cfg.S3PublicBase)
	})

	t.Run("load with custom values", func(t *testing.T) {
		os.Setenv("APP_PORT", "3000")
		os.Setenv("DYNAMODB_TABLE", "custom_documents")
		os.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
		os.Setenv("AWS_ACCESS_KEY_ID", "test_key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test_secret")
		os.Setenv("AWS_REGION", "eu-west-1")
		os.Setenv("S3_BUCKET", "custom_bucket")
		os.Setenv("S3_ENDPOINT", "http://localhost:9000")
		os.Setenv("S3_USE_PATH_STYLE", "true")
		os.Setenv("S3_PUBLIC_BASE_URL", "http://public.example.com")

		cfg := config.Load()

		assert.Equal(t, ":3000", cfg.Port)
		assert.Equal(t, "custom_documents", cfg.DynamoDBTable)
		assert.Equal(t, "http://localhost:8000", cfg.DynamoDBEndpoint)
		assert.Equal(t, "test_key", cfg.AWSAccessKey)
		assert.Equal(t, "test_secret", cfg.AWSSecretKey)
		assert.Equal(t, "eu-west-1", cfg.AWSRegion)
		assert.Equal(t, "custom_bucket", cfg.S3Bucket)
		assert.Equal(t, "http://localhost:9000", cfg.S3Endpoint)
		assert.True(t, cfg.S3UsePath)
		assert.Equal(t, "http://public.example.com", cfg.S3PublicBase)

		// Cleanup
		os.Clearenv()
	})

	t.Run("load with RabbitMQ config", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("RABBITMQ_URL", "amqp://user:pass@rabbitmq:5672/")
		os.Setenv("RABBITMQ_CONSUMER_QUEUE", "custom.queue")
		os.Setenv("RABBITMQ_AUTH_REQUEST_QUEUE", "auth.request")
		os.Setenv("RABBITMQ_AUTH_RESULT_QUEUE", "auth.result")

		cfg := config.Load()

		assert.Equal(t, "amqp://user:pass@rabbitmq:5672/", cfg.RabbitMQ.URL)
		assert.Equal(t, "custom.queue", cfg.RabbitMQ.ConsumerQueue)
		assert.Equal(t, "auth.request", cfg.RabbitMQ.AuthenticationRequestQueue)
		assert.Equal(t, "auth.result", cfg.RabbitMQ.AuthenticationResultQueue)

		os.Clearenv()
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:  "test-bucket",
			AWSRegion: "us-east-1",
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing S3 bucket", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:  "",
			AWSRegion: "us-east-1",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "S3_BUCKET required")
	})

	t.Run("S3 endpoint without path style", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:   "test-bucket",
			S3Endpoint: "http://localhost:9000",
			S3UsePath:  false,
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "S3_USE_PATH_STYLE=true required")
	})

	t.Run("S3 endpoint with path style", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:   "test-bucket",
			S3Endpoint: "http://localhost:9000",
			S3UsePath:  true,
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("AWS S3 without region", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:   "test-bucket",
			S3Endpoint: "",
			AWSRegion:  "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS_REGION required")
	})

	t.Run("MinIO with path style", func(t *testing.T) {
		cfg := &config.Config{
			S3Bucket:   "test-bucket",
			S3Endpoint: "http://minio:9000",
			S3UsePath:  true,
			AWSRegion:  "", // Region not required for MinIO
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
