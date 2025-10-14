package config

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL           string
	
	// Consumer queue configuration
	ConsumerQueue string
	
	// Publisher queue configuration
	AuthenticationRequestQueue string
	
	// Queue settings
	Durable       bool
	PrefetchCount int
	AutoAck       bool
}

// DefaultRabbitMQConfig returns sensible defaults for RabbitMQ
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Durable:       true,  // Queues persist across restarts
		PrefetchCount: 1,     // Process 1 message at a time per consumer
		AutoAck:       false, // Manual acknowledgment for reliability
	}
}
