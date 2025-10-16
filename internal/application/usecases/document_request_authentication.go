package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentRequestAuthenticationService defines the interface for document authentication request operations
type DocumentRequestAuthenticationService interface {
	RequestAuthentication(ctx context.Context, documentID string) error
}

type documentRequestAuthenticationService struct {
	repo          interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
	publisher     interfaces.MessagePublisher
	queue         string
	expiration    time.Duration
}

// NewDocumentRequestAuthenticationService creates a new document authentication request service
func NewDocumentRequestAuthenticationService(
	repo interfaces.DocumentRepository,
	objectStorage interfaces.ObjectStorage,
	publisher interfaces.MessagePublisher,
	queue string,
	expiration time.Duration,
) DocumentRequestAuthenticationService {
	if expiration == 0 {
		expiration = 24 * time.Hour // Default: 24 hours for authentication URLs
	}
	return &documentRequestAuthenticationService{
		repo:          repo,
		objectStorage: objectStorage,
		publisher:     publisher,
		queue:         queue,
		expiration:    expiration,
	}
}

// RequestAuthentication requests authentication for a document by publishing an event
func (s *documentRequestAuthenticationService) RequestAuthentication(
	ctx context.Context,
	documentID string,
) error {
	doc, err := s.repo.GetByID(ctx, documentID)
	if err != nil {
		return err
	}

	if doc == nil {
		return errors.NewNotFoundError(fmt.Sprintf("document with ID %s not found", documentID))
	}

	if err := s.repo.UpdateAuthenticationStatus(ctx, documentID, models.AuthenticationStatusAuthenticating); err != nil {
		return fmt.Errorf("failed to update authentication status: %w", err)
	}

	presignedURL, err := s.objectStorage.GeneratePresignedURL(ctx, doc.ObjectKey, s.expiration)
	if err != nil {
		return fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	event := events.DocumentAuthenticationRequestedEvent{
		IDCitizen:     doc.OwnerID,
		URLDocument:   presignedURL,
		DocumentTitle: doc.Filename,
		DocumentID:    doc.ID,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := s.publisher.Publish(ctx, s.queue, eventJSON); err != nil {
		return fmt.Errorf("failed to publish authentication request event: %w", err)
	}

	return nil
}
