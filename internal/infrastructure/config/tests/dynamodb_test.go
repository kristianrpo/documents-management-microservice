package config_test

import (
	"context"
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDynamoDBClient(t *testing.T) {
	ctx := context.Background()

	t.Run("creates client with valid configuration", func(t *testing.T) {
		client, err := config.NewDynamoDBClient(
			ctx,
			"test_access_key",
			"test_secret_key",
			"us-east-1",
			"",
		)

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client with local endpoint", func(t *testing.T) {
		client, err := config.NewDynamoDBClient(
			ctx,
			"local",
			"local",
			"us-east-1",
			"http://localhost:8000",
		)

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client without explicit credentials", func(t *testing.T) {
		client, err := config.NewDynamoDBClient(
			ctx,
			"",
			"",
			"us-east-1",
			"",
		)

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("creates client with different regions", func(t *testing.T) {
		regions := []string{"us-east-1", "eu-west-1", "ap-southeast-1"}

		for _, region := range regions {
			client, err := config.NewDynamoDBClient(
				ctx,
				"test_key",
				"test_secret",
				region,
				"",
			)

			assert.NoError(t, err)
			assert.NotNil(t, client)
		}
	})
}
