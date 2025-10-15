package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
)

type DocumentRequestAuthenticationService struct {
	repo          interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
	publisher     interfaces.MessagePublisher
	queue         string
	expiration    time.Duration
}

func NewDocumentRequestAuthenticationService(
	repo interfaces.DocumentRepository,
	objectStorage interfaces.ObjectStorage,
	publisher interfaces.MessagePublisher,
	queue string,
	expiration time.Duration,
) *DocumentRequestAuthenticationService {
	if expiration == 0 {
		expiration = 24 * time.Hour // Default: 24 hours for authentication URLs
	}
	return &DocumentRequestAuthenticationService{
		repo:          repo,
		objectStorage: objectStorage,
		publisher:     publisher,
		queue:         queue,
		expiration:    expiration,
	}
}

// RequestAuthentication requests authentication for a document by publishing an event
func (s *DocumentRequestAuthenticationService) RequestAuthentication(
	ctx context.Context,
	documentID string,
) error {
	// Get the document
	doc, err := s.repo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}

	if doc == nil {
		return errors.NewNotFoundError(fmt.Sprintf("document with ID %s not found", documentID))
	}

	// Generate pre-signed URL
	presignedURL, err := s.objectStorage.GeneratePresignedURL(ctx, doc.ObjectKey, s.expiration)
	if err != nil {
		return fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	// Create the event using document information
	// IDCitizen is the owner's ID
	// DocumentTitle is the filename
	event := events.DocumentAuthenticationRequestedEvent{
		IDCitizen:     doc.OwnerID,
		URLDocument:   presignedURL,
		DocumentTitle: doc.Filename,
	}

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish the event
	if err := s.publisher.Publish(ctx, s.queue, eventJSON); err != nil {
		return fmt.Errorf("failed to publish authentication request event: %w", err)
	}

	log.Printf("Authentication requested for document %s (owner ID: %d, title: %s)", 
		documentID, doc.OwnerID, doc.Filename)

	return nil
}
