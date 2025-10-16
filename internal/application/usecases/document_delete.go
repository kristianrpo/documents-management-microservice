package usecases

import (
	"context"
	"fmt"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// DocumentDeleteService defines the interface for deleting individual documents
type DocumentDeleteService interface {
	Delete(ctx context.Context, id string) error
}

type documentDeleteService struct {
	repository    interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
}

// NewDocumentDeleteService creates a new document deletion service
func NewDocumentDeleteService(repository interfaces.DocumentRepository, objectStorage interfaces.ObjectStorage) DocumentDeleteService {
	return &documentDeleteService{
		repository:    repository,
		objectStorage: objectStorage,
	}
}

// Delete removes a document and its associated file from storage
func (s *documentDeleteService) Delete(ctx context.Context, id string) error {
	document, err := s.repository.DeleteByID(ctx, id)
	if err != nil {
		return errors.NewPersistenceError(err)
	}

	if document == nil {
		return errors.NewNotFoundError("document not found")
	}

	if err := s.objectStorage.Delete(ctx, document.ObjectKey); err != nil {
		return fmt.Errorf("failed to delete object from storage (metadata was already deleted): %w", err)
	}

	return nil
}
