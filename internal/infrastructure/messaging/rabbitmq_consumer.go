package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
)

// RabbitMQConsumer implements the MessageConsumer interface for RabbitMQ
type RabbitMQConsumer struct {
	client  *RabbitMQClient
	channel *amqp091.Channel
}

// NewRabbitMQConsumer creates a new RabbitMQ message consumer
func NewRabbitMQConsumer(client *RabbitMQClient) (*RabbitMQConsumer, error) {
	channel, err := client.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer channel: %w", err)
	}

	return &RabbitMQConsumer{
		client:  client,
		channel: channel,
	}, nil
}

// SubscribeToQueue starts consuming messages from the specified queue with the provided handler
func (r *RabbitMQConsumer) SubscribeToQueue(ctx context.Context, queueName string, handler interfaces.MessageHandler) error {
	if err := r.setupConsumer(queueName); err != nil {
		return err
	}

	msgs, err := r.startConsuming(queueName)
	if err != nil {
		return err
	}

	log.Printf("RabbitMQ consumer subscribed to queue: %s", queueName)

	go r.consumeMessages(ctx, queueName, msgs, handler)
	return nil
}

// setupConsumer declares the queue and sets QoS parameters
func (r *RabbitMQConsumer) setupConsumer(queueName string) error {
	cfg := r.client.GetConfig()

	if err := r.client.DeclareQueue(r.channel, queueName); err != nil {
		return err
	}

	err := r.channel.Qos(cfg.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

// startConsuming begins consuming messages from the queue
func (r *RabbitMQConsumer) startConsuming(queueName string) (<-chan amqp091.Delivery, error) {
	cfg := r.client.GetConfig()

	msgs, err := r.channel.Consume(
		queueName,   // queue
		"",          // consumer tag
		cfg.AutoAck, // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer for queue %s: %w", queueName, err)
	}

	return msgs, nil
}

// consumeMessages processes messages in a goroutine
func (r *RabbitMQConsumer) consumeMessages(ctx context.Context, queueName string, msgs <-chan amqp091.Delivery, handler interfaces.MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context canceled, stopping consumer for queue: %s", queueName)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Message channel closed for queue: %s", queueName)
				return
			}
			r.processMessage(ctx, queueName, msg, handler)
		}
	}
}

// processMessage handles a single message with error handling and acknowledgment
func (r *RabbitMQConsumer) processMessage(ctx context.Context, queueName string, msg amqp091.Delivery, handler interfaces.MessageHandler) {
	err := handler(ctx, msg.Body)
	if err != nil {
		log.Printf("Error processing message from queue %s: %v", queueName, err)
		if nackErr := msg.Nack(false, true); nackErr != nil {
			log.Printf("Failed to NACK message: %v", nackErr)
		}
		return
	}

	cfg := r.client.GetConfig()
	if !cfg.AutoAck {
		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("Failed to ACK message: %v", ackErr)
		}
	}
}

// Close closes the consumer channel (connection is managed by RabbitMQClient)
func (r *RabbitMQConsumer) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("error closing channel: %v", err)
		}
	}
	return nil
}
