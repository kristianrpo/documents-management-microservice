package usecases

import (
	"context"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentGetService defines the interface for retrieving individual documents
type DocumentGetService interface {
	GetByID(ctx context.Context, id string) (*models.Document, error)
}

type documentGetService struct {
	repository interfaces.DocumentRepository
	storage    interfaces.ObjectStorage
}

// NewDocumentGetService creates a new document retrieval service
// storage is used to generate pre-signed URLs for document access when fetching details
func NewDocumentGetService(repository interfaces.DocumentRepository, storage interfaces.ObjectStorage) DocumentGetService {
	return &documentGetService{
		repository: repository,
		storage:    storage,
	}
}

// GetByID retrieves a document by its unique identifier and populates a pre-signed URL
func (s *documentGetService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	document, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewPersistenceError(err)
	}

	if document == nil {
		return nil, errors.NewNotFoundError("document not found")
	}

	// If we have an object storage provider, generate a presigned URL for the object's key
	if s.storage != nil && document.ObjectKey != "" {
		// Choose a reasonable default expiration for pre-signed URLs
		presigned, perr := s.storage.GeneratePresignedURL(ctx, document.ObjectKey, 15*time.Minute)
		if perr != nil {
			return nil, errors.NewPersistenceError(perr)
		}
		document.URL = presigned
	}

	return document, nil
}
