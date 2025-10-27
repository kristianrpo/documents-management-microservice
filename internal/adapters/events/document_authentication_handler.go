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
	repo             interfaces.DocumentRepository
	processedMsgRepo interfaces.ProcessedMessageRepository
}

// NewDocumentAuthenticationHandler creates a new handler for document authentication events
func NewDocumentAuthenticationHandler(repo interfaces.DocumentRepository, processedMsgRepo interfaces.ProcessedMessageRepository) *DocumentAuthenticationHandler {
	return &DocumentAuthenticationHandler{
		repo:             repo,
		processedMsgRepo: processedMsgRepo,
	}
}

// HandleAuthenticationCompleted processes the document authentication completed event
func (h *DocumentAuthenticationHandler) HandleAuthenticationCompleted(ctx context.Context, message []byte) error {
	var event events.DocumentAuthenticationCompletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal authentication completed event: %w", err)
	}

	log.Printf("processing authentication completed event for document ID: %s, messageId: %s, citizen ID: %d, authenticated: %v",
		event.DocumentID, event.MessageID, event.IDCitizen, event.Authenticated)

	// IDEMPOTENCIA: Verificar si el mensaje ya fue procesado usando DynamoDB
	if event.MessageID != "" && h.processedMsgRepo != nil {
		alreadyProcessed, err := h.processedMsgRepo.CheckIfProcessed(ctx, event.MessageID)
		if err != nil {
			log.Printf("warning: failed to check if message is processed: %v, continuing with processing", err)
		} else if alreadyProcessed {
			log.Printf("message %s already processed, skipping (idempotent)", event.MessageID)
			return nil // Idempotente: no error, solo skip
		}
	}

	// Procesar el mensaje
	var newStatus models.AuthenticationStatus
	if event.Authenticated {
		newStatus = models.AuthenticationStatusAuthenticated
	} else {
		newStatus = models.AuthenticationStatusUnauthenticated
	}

	// Actualizar el estado del documento
	if err := h.repo.UpdateAuthenticationStatus(ctx, event.DocumentID, newStatus); err != nil {
		return fmt.Errorf("failed to update document authentication status: %w", err)
	}

	log.Printf("Document status updated to %s for document %s", newStatus, event.DocumentID)

	// Marcar el mensaje como procesado para idempotencia
	// IMPORTANTE: Solo hacemos ACK después de procesar completamente
	// Si esto falla, el mensaje será re-enviado (idempotente gracias al check anterior)
	if event.MessageID != "" && h.processedMsgRepo != nil {
		processedMsg := models.NewProcessedMessage(event.MessageID, event.DocumentID, "authentication-handler")
		if err := h.processedMsgRepo.MarkAsProcessed(ctx, processedMsg); err != nil {
			log.Printf("error: failed to mark message as processed: %v", err)
			return fmt.Errorf("failed to mark message as processed: %w", err)
		}
		log.Printf("Message %s marked as processed", event.MessageID)
	}

	log.Printf("successfully processed authentication event for document %s (messageId: %s, status: %s, message: %s)",
		event.DocumentID, event.MessageID, newStatus, event.Message)

	return nil // ACK del mensaje
}
