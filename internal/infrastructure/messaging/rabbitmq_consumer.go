package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
)

type RabbitMQConsumer struct {
	client  *RabbitMQClient
	channel *amqp091.Channel
}

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

func (r *RabbitMQConsumer) Subscribe(ctx context.Context, handler interfaces.MessageHandler) error {
	cfg := r.client.GetConfig()
	
	// Declare the queue (idempotent operation)
	if err := r.client.DeclareQueue(r.channel, cfg.ConsumerQueue); err != nil {
		return err
	}

	// Set QoS (quality of service)
	err := r.channel.Qos(
		cfg.PrefetchCount, // prefetch count
		0,                 // prefetch size
		false,             // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := r.channel.Consume(
		cfg.ConsumerQueue, // queue
		"",                // consumer tag
		cfg.AutoAck,       // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("RabbitMQ consumer subscribed to queue: %s", cfg.ConsumerQueue)

	// Process messages in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping consumer")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				// Process the message
				err := handler(ctx, msg.Body)
				if err != nil {
				log.Printf("Error processing message: %v", err)
				// NACK the message (requeue it)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("Failed to NACK message: %v", nackErr)
				}
			} else {
				// ACK the message
				cfg := r.client.GetConfig()
				if !cfg.AutoAck {
					if ackErr := msg.Ack(false); ackErr != nil {
						log.Printf("Failed to ACK message: %v", ackErr)
					}
				}
			}
		}
	}
}()
	return nil
}

func (r *RabbitMQConsumer) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("error closing channel: %v", err)
		}
	}
	// Connection is managed by RabbitMQClient, only close the channel
	return nil
}
