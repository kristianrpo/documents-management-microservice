package interfaces

import (
	"context"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// ProcessedMessageRepository defines the interface for managing processed messages (idempotency)
type ProcessedMessageRepository interface {
	// CheckIfProcessed checks if a message with the given messageID has been processed
	CheckIfProcessed(ctx context.Context, messageID string) (bool, error)

	// MarkAsProcessed marks a message as processed
	MarkAsProcessed(ctx context.Context, message *models.ProcessedMessage) error
}
