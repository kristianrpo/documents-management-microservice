package config_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRabbitMQConfig(t *testing.T) {
	t.Run("returns config with sensible defaults", func(t *testing.T) {
		cfg := config.DefaultRabbitMQConfig()

		assert.True(t, cfg.Durable, "queues should be durable by default")
		assert.Equal(t, 1, cfg.PrefetchCount, "should prefetch 1 message at a time")
		assert.False(t, cfg.AutoAck, "should use manual acknowledgment by default")
	})

	t.Run("allows custom values", func(t *testing.T) {
		cfg := config.DefaultRabbitMQConfig()
		
		cfg.URL = "amqp://user:pass@localhost:5672/"
		cfg.ConsumerQueue = "my.queue"
		cfg.AuthenticationRequestQueue = "auth.request"
		cfg.AuthenticationResultQueue = "auth.result"
		cfg.PrefetchCount = 5

		assert.Equal(t, "amqp://user:pass@localhost:5672/", cfg.URL)
		assert.Equal(t, "my.queue", cfg.ConsumerQueue)
		assert.Equal(t, "auth.request", cfg.AuthenticationRequestQueue)
		assert.Equal(t, "auth.result", cfg.AuthenticationResultQueue)
		assert.Equal(t, 5, cfg.PrefetchCount)
	})
}

func TestRabbitMQConfig_Structure(t *testing.T) {
	t.Run("config has all required fields", func(t *testing.T) {
		cfg := config.RabbitMQConfig{
			URL:                        "amqp://localhost:5672/",
			ConsumerQueue:              "consumer.queue",
			AuthenticationRequestQueue: "auth.request.queue",
			AuthenticationResultQueue:  "auth.result.queue",
			Durable:                    true,
			PrefetchCount:              10,
			AutoAck:                    false,
		}

		assert.Equal(t, "amqp://localhost:5672/", cfg.URL)
		assert.Equal(t, "consumer.queue", cfg.ConsumerQueue)
		assert.Equal(t, "auth.request.queue", cfg.AuthenticationRequestQueue)
		assert.Equal(t, "auth.result.queue", cfg.AuthenticationResultQueue)
		assert.True(t, cfg.Durable)
		assert.Equal(t, 10, cfg.PrefetchCount)
		assert.False(t, cfg.AutoAck)
	})
}
