package usecases

import (
	"context"
	"log"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

type DocumentDeleteService interface {
	Delete(ctx context.Context, id string) error
}

type documentDeleteService struct {
	repository    interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
}

func NewDocumentDeleteService(repository interfaces.DocumentRepository, objectStorage interfaces.ObjectStorage) DocumentDeleteService {
	return &documentDeleteService{
		repository:    repository,
		objectStorage: objectStorage,
	}
}

func (s *documentDeleteService) Delete(ctx context.Context, id string) error {
	document, err := s.repository.DeleteByID(ctx, id)
	if err != nil {
		return errors.NewPersistenceError(err)
	}

	if document == nil {
		return errors.NewNotFoundError("document not found")
	}

	if err := s.objectStorage.Delete(ctx, document.ObjectKey); err != nil {
		log.Printf("failed to delete object from S3: %v (document metadata was already deleted)", err)
	}

	return nil
}
