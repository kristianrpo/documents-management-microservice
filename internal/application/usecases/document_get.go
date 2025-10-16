package usecases

import (
	"context"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// DocumentGetService defines the interface for retrieving individual documents
type DocumentGetService interface {
	GetByID(ctx context.Context, id string) (*models.Document, error)
}

type documentGetService struct {
	repository interfaces.DocumentRepository
}

// NewDocumentGetService creates a new document retrieval service
func NewDocumentGetService(repository interfaces.DocumentRepository) DocumentGetService {
	return &documentGetService{
		repository: repository,
	}
}

// GetByID retrieves a document by its unique identifier
func (s *documentGetService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	document, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewPersistenceError(err)
	}

	if document == nil {
		return nil, errors.NewNotFoundError("document not found")
	}

	return document, nil
}
