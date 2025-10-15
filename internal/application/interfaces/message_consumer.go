package interfaces

import "context"

// MessageHandler defines the function signature for message handlers
type MessageHandler func(ctx context.Context, message []byte) error

// MessageConsumer defines the interface for consuming messages from message queues (RabbitMQ, Kafka, SQS, etc.)
type MessageConsumer interface {
	// Subscribe starts consuming messages from a queue and processes them with the provided handler
	Subscribe(ctx context.Context, handler MessageHandler) error

	// Close closes the connection to the message broker
	Close() error
}
