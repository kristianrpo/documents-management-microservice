package interfaces

import "context"

// MessagePublisher defines the interface for publishing messages to message queues
type MessagePublisher interface {
	// Publish sends a message to the specified queue/exchange
	Publish(ctx context.Context, queue string, message []byte) error

	// Close closes the connection to the message broker
	Close() error
}
