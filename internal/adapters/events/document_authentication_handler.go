package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentAuthenticationHandler handles authentication completion events
type DocumentAuthenticationHandler struct {
	repo interfaces.DocumentRepository
}

// NewDocumentAuthenticationHandler creates a new handler for document authentication events
func NewDocumentAuthenticationHandler(repo interfaces.DocumentRepository) *DocumentAuthenticationHandler {
	return &DocumentAuthenticationHandler{
		repo: repo,
	}
}

// HandleAuthenticationCompleted processes the document authentication completed event
func (h *DocumentAuthenticationHandler) HandleAuthenticationCompleted(ctx context.Context, message []byte) error {
	var event events.DocumentAuthenticationCompletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal authentication completed event: %w", err)
	}

	log.Printf("processing authentication completed event for document ID: %s, citizen ID: %d, authenticated: %v",
		event.DocumentID, event.IDCitizen, event.Authenticated)

	var newStatus models.AuthenticationStatus
	if event.Authenticated {
		newStatus = models.AuthenticationStatusAuthenticated
	} else {
		newStatus = models.AuthenticationStatusUnauthenticated
	}

	if err := h.repo.UpdateAuthenticationStatus(ctx, event.DocumentID, newStatus); err != nil {
		return fmt.Errorf("failed to update document authentication status: %w", err)
	}

	log.Printf("successfully updated authentication status to %s for document %s (message: %s)",
		newStatus, event.DocumentID, event.Message)

	return nil
}
