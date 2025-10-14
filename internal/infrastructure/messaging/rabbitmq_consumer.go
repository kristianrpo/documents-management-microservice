package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
)

type RabbitMQConsumer struct {
	conn   *amqp091.Connection
	channel *amqp091.Channel
	config  config.RabbitMQConfig
}

func NewRabbitMQConsumer(cfg config.RabbitMQConfig) (*RabbitMQConsumer, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConsumer{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

func (r *RabbitMQConsumer) Subscribe(ctx context.Context, handler interfaces.MessageHandler) error {
	// Declare the queue (idempotent operation)
	queue, err := r.channel.QueueDeclare(
		r.config.Queue,   // name
		r.config.Durable, // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS (quality of service)
	err = r.channel.Qos(
		r.config.PrefetchCount, // prefetch count
		0,                      // prefetch size
		false,                  // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming messages
	msgs, err := r.channel.Consume(
		queue.Name,        // queue
		"",                // consumer tag
		r.config.AutoAck,  // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("RabbitMQ consumer subscribed to queue: %s", r.config.Queue)

	// Process messages in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping consumer")
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
					if !r.config.AutoAck {
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
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("error closing connection: %v", err)
			return err
		}
	}
	return nil
}
