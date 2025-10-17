package tests

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/messaging"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQPublisher_Close_NoChannel(t *testing.T) {
	pub := &messaging.RabbitMQPublisher{}
	err := pub.Close()
	assert.NoError(t, err)
}

func TestRabbitMQConsumer_Close_NoChannel(t *testing.T) {
	cons := &messaging.RabbitMQConsumer{}
	err := cons.Close()
	assert.NoError(t, err)
}
