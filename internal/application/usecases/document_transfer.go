package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

const (
	// maxTransferDocuments defines the maximum number of documents that can be transferred in one operation
	maxTransferDocuments = 1000
)

// DocumentTransferResult represents a document with its pre-signed URL
type DocumentTransferResult struct {
	Document     *models.Document
	PresignedURL string
	ExpiresAt    time.Time
}

// DocumentTransferService defines the interface for document transfer operations
type DocumentTransferService interface {
	PrepareTransfer(ctx context.Context, ownerID int64) ([]DocumentTransferResult, error)
}

type documentTransferService struct {
	repo          interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
	expiration    time.Duration
}

// NewDocumentTransferService creates a new document transfer service
func NewDocumentTransferService(
	repo interfaces.DocumentRepository,
	objectStorage interfaces.ObjectStorage,
	expiration time.Duration,
) DocumentTransferService {
	if expiration == 0 {
		expiration = 15 * time.Minute // Default: 15 minutes
	}
	return &documentTransferService{
		repo:          repo,
		objectStorage: objectStorage,
		expiration:    expiration,
	}
}

// PrepareTransfer generates pre-signed URLs for all documents owned by a user
func (s *documentTransferService) PrepareTransfer(ctx context.Context, ownerID int64) ([]DocumentTransferResult, error) {
	// List all documents for the user
	documents, _, err := s.repo.List(ctx, ownerID, maxTransferDocuments, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	if len(documents) == 0 {
		return []DocumentTransferResult{}, nil
	}

	results := make([]DocumentTransferResult, 0, len(documents))
	expiresAt := time.Now().Add(s.expiration)

	for _, doc := range documents {
		presignedURL, err := s.objectStorage.GeneratePresignedURL(ctx, doc.ObjectKey, s.expiration)
		if err != nil {
			return nil, fmt.Errorf("failed to generate pre-signed URL for document %s: %w", doc.ID, err)
		}

		results = append(results, DocumentTransferResult{
			Document:     doc,
			PresignedURL: presignedURL,
			ExpiresAt:    expiresAt,
		})
	}

	return results, nil
}
