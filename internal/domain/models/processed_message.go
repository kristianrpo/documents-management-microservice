package models

import "time"

// ProcessedMessage represents a message that has been processed (for idempotency)
type ProcessedMessage struct {
	MessageID   string    `json:"messageId"`   // Primary key: unique message identifier
	ProcessedAt time.Time `json:"processedAt"` // Timestamp when the message was processed
	DocumentID  string    `json:"documentId"`  // Related document ID (for reference)
	ProcessedBy string    `json:"processedBy"` // Service/handler that processed the message
	TTL         int64     `json:"ttl"`         // DynamoDB TTL attribute (Unix timestamp)
}

// NewProcessedMessage creates a new ProcessedMessage with default TTL of 7 days
func NewProcessedMessage(messageID, documentID, processedBy string) *ProcessedMessage {
	now := time.Now()
	ttl := now.Add(7 * 24 * time.Hour).Unix() // 7 days TTL

	return &ProcessedMessage{
		MessageID:   messageID,
		ProcessedAt: now,
		DocumentID:  documentID,
		ProcessedBy: processedBy,
		TTL:         ttl,
	}
}
