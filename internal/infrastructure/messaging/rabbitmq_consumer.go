package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
)

// RabbitMQConsumer implements the MessageConsumer interface for RabbitMQ
type RabbitMQConsumer struct {
	client        *RabbitMQClient
	channel       *amqp091.Channel
	subscriptions map[string]interfaces.MessageHandler // queueName -> handler
}

// NewRabbitMQConsumer creates a new RabbitMQ message consumer
func NewRabbitMQConsumer(client *RabbitMQClient) (*RabbitMQConsumer, error) {
	// Attempt to create a channel, but do not fail if RabbitMQ is down.
	channel, _ := client.CreateChannel()

	r := &RabbitMQConsumer{
		client:        client,
		channel:       channel, // may be nil
		subscriptions: make(map[string]interfaces.MessageHandler),
	}

	// Start monitor for channel close and initial creation/re-subscription
	go r.monitorChannel()

	return r, nil
}

// SubscribeToQueue starts consuming messages from the specified queue with the provided handler
func (r *RabbitMQConsumer) SubscribeToQueue(ctx context.Context, queueName string, handler interfaces.MessageHandler) error {
	// Keep subscription for re-subscribe after reconnect
	if r.subscriptions == nil {
		r.subscriptions = make(map[string]interfaces.MessageHandler)
	}
	r.subscriptions[queueName] = handler
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

	if r.channel == nil || r.channel.IsClosed() {
		// Attempt to get a channel lazily
		ch, err := r.client.CreateChannel()
		if err != nil {
			return fmt.Errorf("consumer channel unavailable: %w", err)
		}
		r.channel = ch
	}

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
				log.Printf("Message channel closed for queue: %s (will attempt to resubscribe)", queueName)
				return
			}
			r.processMessage(ctx, queueName, msg, handler)
		}
	}
}

// monitorChannel watches the consumer channel and re-creates it on close, then re-subscribes
func (r *RabbitMQConsumer) monitorChannel() {
	for {
		ch := r.channel
		if ch == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		closeCh := ch.NotifyClose(make(chan *amqp091.Error))
		if err := <-closeCh; err != nil {
			log.Printf("Consumer channel closed: %v. Reconnecting...", err)
		} else {
			log.Printf("Consumer channel closed cleanly")
			return
		}

		// Try to recreate channel until success
		for {
			time.Sleep(2 * time.Second)
			newCh, err := r.client.CreateChannel()
			if err != nil {
				log.Printf("Failed to recreate consumer channel: %v", err)
				continue
			}
			r.channel = newCh
			log.Printf("Consumer channel recreated successfully")
			break
		}

		// Resubscribe to all queues
		for q, h := range r.subscriptions {
			if err := r.setupConsumer(q); err != nil {
				log.Printf("Failed to setup consumer for queue %s after reconnect: %v", q, err)
				continue
			}
			msgs, err := r.startConsuming(q)
			if err != nil {
				log.Printf("Failed to restart consuming for queue %s: %v", q, err)
				continue
			}
			go r.consumeMessages(context.Background(), q, msgs, h)
			log.Printf("Resubscribed to queue: %s", q)
		}
	}
}

// processMessage handles a single message with error handling and acknowledgment
func (r *RabbitMQConsumer) processMessage(ctx context.Context, queueName string, msg amqp091.Delivery, handler interfaces.MessageHandler) {
	// Extract MessageID from headers for logging and potential deduplication
	messageID := ""
	if msg.MessageId != "" {
		messageID = msg.MessageId
	} else if msg.Headers != nil {
		if id, ok := msg.Headers["x-message-id"].(string); ok {
			messageID = id
		}
	}

	if messageID != "" {
		log.Printf("Processing message from queue %s with messageId: %s", queueName, messageID)
	}

	// Process the message with the handler
	err := handler(ctx, msg.Body)
	if err != nil {
		// Handler returned an error - NACK the message to requeue it
		log.Printf("Error processing message from queue %s (messageId: %s): %v", queueName, messageID, err)
		log.Printf("Message will be NACK'd and requeued for retry")
		if nackErr := msg.Nack(false, true); nackErr != nil {
			log.Printf("Failed to NACK message: %v", nackErr)
		} else {
			log.Printf("   ✓ Message NACK'd - will be retried")
		}
		return
	}

	// Handler returned nil (success) - ACK the message
	cfg := r.client.GetConfig()
	if !cfg.AutoAck {
		if ackErr := msg.Ack(false); ackErr != nil {
			log.Printf("Failed to ACK message: %v", ackErr)
		} else {
			log.Printf("Message ACK'd successfully (messageId: %s, queue: %s)", messageID, queueName)
		}
	}

	log.Printf("Successfully processed and acknowledged message with messageId: %s from queue %s", messageID, queueName)
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
