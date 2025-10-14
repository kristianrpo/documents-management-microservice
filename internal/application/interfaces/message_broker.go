package interfaces

import "context"

// MessageHandler defines the function signature for message handlers
type MessageHandler func(ctx context.Context, message []byte) error

// MessageBroker defines the interface for message queue systems (RabbitMQ, Kafka, SQS, etc.)
type MessageBroker interface {
	// Subscribe starts consuming messages from a queue and processes them with the provided handler
	Subscribe(ctx context.Context, handler MessageHandler) error

	// Close closes the connection to the message broker
	Close() error
}
