package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

// DocumentTransferResult represents a document with its pre-signed URL
type DocumentTransferResult struct {
	Document     *domain.Document
	PresignedURL string
	ExpiresAt    time.Time
}

type DocumentTransferService struct {
	repo          interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
	expiration    time.Duration
}

func NewDocumentTransferService(
	repo interfaces.DocumentRepository,
	objectStorage interfaces.ObjectStorage,
	expiration time.Duration,
) *DocumentTransferService {
	if expiration == 0 {
		expiration = 15 * time.Minute // Default: 15 minutes
	}
	return &DocumentTransferService{
		repo:          repo,
		objectStorage: objectStorage,
		expiration:    expiration,
	}
}

// PrepareTransfer generates pre-signed URLs for all documents owned by a user
func (s *DocumentTransferService) PrepareTransfer(ctx context.Context, ownerID int64) ([]DocumentTransferResult, error) {
	// List all documents for the user
	documents, _, err := s.repo.List(ctx, ownerID, 1000, 0) // Get up to 1000 documents
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	if len(documents) == 0 {
		return []DocumentTransferResult{}, nil
	}

	results := make([]DocumentTransferResult, 0, len(documents))
	expiresAt := time.Now().Add(s.expiration)

	for _, doc := range documents {
		// Generate pre-signed URL for each document
		presignedURL, err := s.objectStorage.GeneratePresignedURL(ctx, doc.ObjectKey, s.expiration)
		if err != nil {
			log.Printf("Warning: failed to generate pre-signed URL for document %s: %v", doc.ID, err)
			continue // Skip this document but continue with others
		}

		results = append(results, DocumentTransferResult{
			Document:     doc,
			PresignedURL: presignedURL,
			ExpiresAt:    expiresAt,
		})
	}

	return results, nil
}
