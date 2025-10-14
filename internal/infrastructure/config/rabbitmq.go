package config

// RabbitMQConfig holds RabbitMQ consumer configuration
type RabbitMQConfig struct {
	URL           string
	Queue         string
	Durable       bool
	PrefetchCount int
	AutoAck       bool
}

// DefaultRabbitMQConfig returns sensible defaults for RabbitMQ consumer
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Durable:       true,  // Queues persist across restarts
		PrefetchCount: 1,     // Process 1 message at a time per consumer
		AutoAck:       false, // Manual acknowledgment for reliability
	}
}
