package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements the MessagePublisher interface for RabbitMQ
type RabbitMQPublisher struct {
	client  *RabbitMQClient
	channel *amqp091.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ message publisher
func NewRabbitMQPublisher(client *RabbitMQClient) (*RabbitMQPublisher, error) {
	// Lazy initialization: do not require a channel at startup.
	// The channel will be created on-demand in Publish(), allowing the app
	// to start even if RabbitMQ is unavailable initially.
	return &RabbitMQPublisher{
		client:  client,
		channel: nil,
	}, nil
}

// Publish sends a message to the specified RabbitMQ queue
func (p *RabbitMQPublisher) Publish(ctx context.Context, queue string, message []byte) error {
	// Retry a few times in case the connection/channel is being re-established
	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(1 * time.Second)
		}

		ch := p.channel
		if ch == nil || ch.IsClosed() {
			// Try to create a fresh channel
			newCh, err := p.client.CreateChannel()
			if err != nil {
				lastErr = fmt.Errorf("publisher channel unavailable: %w", err)
				continue
			}
			p.channel = newCh
			ch = newCh
		}

		// Declare the queue (idempotent)
		if err := p.client.DeclareQueue(ch, queue); err != nil {
			lastErr = err
			// Force channel refresh on next attempt
			_ = ch.Close()
			p.channel = nil
			continue
		}

		// Extract message ID from the message if present
		var messageID string
		var messageData map[string]interface{}
		if err := json.Unmarshal(message, &messageData); err == nil {
			if id, ok := messageData["messageId"].(string); ok {
				messageID = id
			}
		}

		// Prepare headers for message deduplication
		headers := amqp091.Table{}
		if messageID != "" {
			// Set message_id header for RabbitMQ deduplication
			headers["x-message-id"] = messageID
		}
		// Set timestamp for message ordering
		headers["x-timestamp"] = time.Now().Unix()

		// Publish the message
		err := ch.PublishWithContext(
			ctx,
			"",
			queue,
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				ContentType:  "application/json",
				Body:         message,
				Headers:      headers,
				MessageId:    messageID, // Standard AMQP message ID for deduplication
			},
		)
		if err != nil {
			lastErr = fmt.Errorf("failed to publish message: %w", err)
			// Force channel refresh on next attempt
			_ = ch.Close()
			p.channel = nil
			continue
		}

		log.Printf("Published message to queue: %s (messageId: %s)", queue, messageID)
		return nil
	}
	return fmt.Errorf("publish failed after retries: %w", lastErr)
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
