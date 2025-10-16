package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements the MessagePublisher interface for RabbitMQ
type RabbitMQPublisher struct {
	client  *RabbitMQClient
	channel *amqp091.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ message publisher
func NewRabbitMQPublisher(client *RabbitMQClient) (*RabbitMQPublisher, error) {
	channel, err := client.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher channel: %w", err)
	}

	return &RabbitMQPublisher{
		client:  client,
		channel: channel,
	}, nil
}

// Publish sends a message to the specified RabbitMQ queue
func (p *RabbitMQPublisher) Publish(ctx context.Context, queue string, message []byte) error {
	// Declare the queue (idempotent operation)
	if err := p.client.DeclareQueue(p.channel, queue); err != nil {
		return err
	}

	// Publish the message
	err := p.channel.PublishWithContext(
		ctx,
		"",    // exchange (empty = default exchange)
		queue, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent, // Make message persistent
			ContentType:  "application/json",
			Body:         message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to queue: %s", queue)
	return nil
}

// Close closes the publisher channel (connection is managed by RabbitMQClient)
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ publisher channel: %v", err)
			return err
		}
	}
	// Connection is managed by RabbitMQClient
	return nil
}
