package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

// UserTransferredEvent represents the event payload when a user is transferred
type UserTransferredEvent struct {
	UserEmail string `json:"user_email"`
	EventType string `json:"event_type"`
	Timestamp string `json:"timestamp"`
}

type UserTransferHandler struct {
	deleteAllService usecases.DocumentDeleteAllService
}

func NewUserTransferHandler(deleteAllService usecases.DocumentDeleteAllService) *UserTransferHandler {
	return &UserTransferHandler{
		deleteAllService: deleteAllService,
	}
}

// HandleUserTransferred processes user transfer events and deletes all associated documents
func (h *UserTransferHandler) HandleUserTransferred(ctx context.Context, message []byte) error {
	var event UserTransferredEvent
	
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("failed to unmarshal user transfer event: %v", err)
		return err
	}

	log.Printf("processing user transfer event for email: %s", event.UserEmail)

	// Reuse the existing delete all service
	deletedCount, err := h.deleteAllService.DeleteAll(ctx, event.UserEmail)
	if err != nil {
		log.Printf("failed to delete documents for user %s: %v", event.UserEmail, err)
		return err
	}

	log.Printf("successfully deleted %d documents for user %s", deletedCount, event.UserEmail)
	return nil
}
