package messaging

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
)

// RabbitMQClient manages a shared RabbitMQ connection with auto-reconnection
// Following RabbitMQ best practices: one connection, multiple channels
type RabbitMQClient struct {
	conn          *amqp091.Connection
	config        config.RabbitMQConfig
	mu            sync.RWMutex
	reconnecting  bool
	stopMonitor   chan struct{}
}

// NewRabbitMQClient creates a new RabbitMQ client with a shared connection
// Implements retry logic to handle RabbitMQ startup delays
func NewRabbitMQClient(cfg config.RabbitMQConfig) (*RabbitMQClient, error) {
	var conn *amqp091.Connection
	var err error

	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp091.DialConfig(cfg.URL, amqp091.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})
		if err == nil {
			log.Printf("Connected to RabbitMQ at %s (attempt %d/%d)", cfg.URL, i+1, maxRetries)
			break
		}

		if i < maxRetries-1 {
			log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v. Retrying in %v...",
				i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	c := &RabbitMQClient{
		conn:        conn, // may be nil if initial attempts failed
		config:      cfg,
		stopMonitor: make(chan struct{}),
	}

	if err != nil {
		// Do not fail startup: keep retrying in background until RabbitMQ is available
		log.Printf("RabbitMQ not available at startup: %v. Will keep retrying in background.", err)
		go c.reconnect()
	}

	// Monitor connection and auto-reconnect if it closes
	go c.monitorConnection()

	return c, nil
}

// monitorConnection watches for connection errors and reconnects automatically
func (c *RabbitMQClient) monitorConnection() {
	for {
		c.mu.RLock()
		conn := c.conn
		reconnecting := c.reconnecting
		c.mu.RUnlock()

		if conn == nil {
			if !reconnecting {
				go c.reconnect()
			}
			// Wait a bit and check again until a connection is established or stop is requested
			select {
			case <-time.After(2 * time.Second):
				continue
			case <-c.stopMonitor:
				return
			}
		}

		closeChan := conn.NotifyClose(make(chan *amqp091.Error))

		select {
		case err := <-closeChan:
			if err != nil {
				log.Printf("RabbitMQ connection closed: %v. Attempting to reconnect...", err)
				c.reconnect()
				// continue loop to watch the (eventually) new connection
			} else {
				return
			}
		case <-c.stopMonitor:
			return
		}
	}
}

// reconnect attempts to re-establish connection to RabbitMQ with infinite retries
func (c *RabbitMQClient) reconnect() {
	c.mu.Lock()
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	retryDelay := 5 * time.Second
	attempt := 0
	for {
		attempt++
		conn, err := amqp091.DialConfig(c.config.URL, amqp091.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})
		if err == nil {
			c.mu.Lock()
			// swap connection
			if c.conn != nil && !c.conn.IsClosed() {
				_ = c.conn.Close()
			}
			c.conn = conn
			c.mu.Unlock()
			log.Printf("Successfully reconnected to RabbitMQ (attempt %d)", attempt)
			return
		}
		log.Printf("Reconnection attempt %d failed: %v. Retrying in %v...", attempt, err, retryDelay)
		time.Sleep(retryDelay)
	}
}

// CreateChannel creates a new channel from the shared connection.
// If the connection is closed, it waits for up to 30s for reconnection.
func (c *RabbitMQClient) CreateChannel() (*amqp091.Channel, error) {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error

	for {
		c.mu.RLock()
		conn := c.conn
		reconnecting := c.reconnecting
		c.mu.RUnlock()

		if conn == nil || conn.IsClosed() {
			if !reconnecting {
				go c.reconnect()
			}
			if time.Now().After(deadline) {
				if lastErr == nil {
					lastErr = fmt.Errorf("connection is closed")
				}
				return nil, fmt.Errorf("timeout waiting for RabbitMQ reconnection: %w", lastErr)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			lastErr = fmt.Errorf("failed to create channel: %w", err)
			if time.Now().After(deadline) {
				return nil, lastErr
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return ch, nil
	}
}

// DeclareQueue declares a queue (idempotent operation)
func (c *RabbitMQClient) DeclareQueue(channel *amqp091.Channel, queueName string) error {
	_, err := channel.QueueDeclare(
		queueName,        // name
		c.config.Durable, // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
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

// Close closes the RabbitMQ connection and stops monitoring
func (c *RabbitMQClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.stopMonitor)

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
