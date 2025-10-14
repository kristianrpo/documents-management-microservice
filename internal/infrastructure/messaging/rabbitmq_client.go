package messaging

import (
	"fmt"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
)

// RabbitMQClient manages a shared RabbitMQ connection
// Following RabbitMQ best practices: one connection, multiple channels
type RabbitMQClient struct {
	conn   *amqp091.Connection
	config config.RabbitMQConfig
	mu     sync.RWMutex
}

// NewRabbitMQClient creates a new RabbitMQ client with a shared connection
func NewRabbitMQClient(cfg config.RabbitMQConfig) (*RabbitMQClient, error) {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	log.Printf("Connected to RabbitMQ at %s", cfg.URL)

	return &RabbitMQClient{
		conn:   conn,
		config: cfg,
	}, nil
}

// CreateChannel creates a new channel from the shared connection
func (c *RabbitMQClient) CreateChannel() (*amqp091.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil, fmt.Errorf("connection is closed")
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// DeclareQueue declares a queue (idempotent operation)
func (c *RabbitMQClient) DeclareQueue(channel *amqp091.Channel, queueName string) error {
	_, err := channel.QueueDeclare(
		queueName,       // name
		c.config.Durable, // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}
	return nil
}

// GetConfig returns the RabbitMQ configuration
func (c *RabbitMQClient) GetConfig() config.RabbitMQConfig {
	return c.config
}

// Close closes the RabbitMQ connection
func (c *RabbitMQClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
			return err
		}
		log.Println("RabbitMQ connection closed")
	}
	return nil
}

// IsClosed returns whether the connection is closed
func (c *RabbitMQClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn == nil || c.conn.IsClosed()
}
