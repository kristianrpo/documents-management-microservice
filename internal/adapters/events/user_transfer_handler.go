package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
)

// UserTransferHandler handles user transfer events by deleting all documents owned by the transferred user
type UserTransferHandler struct {
	deleteAllService usecases.DocumentDeleteAllService
}

// NewUserTransferHandler creates a new handler for user transfer events
func NewUserTransferHandler(deleteAllService usecases.DocumentDeleteAllService) *UserTransferHandler {
	return &UserTransferHandler{
		deleteAllService: deleteAllService,
	}
}

// HandleUserTransferred processes user transfer events and deletes all associated documents
func (h *UserTransferHandler) HandleUserTransferred(ctx context.Context, message []byte) error {
	var event events.UserTransferredEvent

	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("failed to unmarshal user transfer event: %v", err)
		return err
	}

	log.Printf("processing user transfer event for citizen ID: %d", event.IDCitizen)

	deletedCount, err := h.deleteAllService.DeleteAll(ctx, event.IDCitizen)
	if err != nil {
		log.Printf("failed to delete documents for user %d: %v", event.IDCitizen, err)
		return err
	}

	log.Printf("successfully deleted %d documents for user %d", deletedCount, event.IDCitizen)
	return nil
}
