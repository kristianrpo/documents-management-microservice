package usecases

import (
	"context"
	"log"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

type DocumentDeleteAllService interface {
	DeleteAll(ctx context.Context, ownerID int64) (int, error)
}

type documentDeleteAllService struct {
	repository    interfaces.DocumentRepository
	objectStorage interfaces.ObjectStorage
}

func NewDocumentDeleteAllService(repository interfaces.DocumentRepository, objectStorage interfaces.ObjectStorage) DocumentDeleteAllService {
	return &documentDeleteAllService{
		repository:    repository,
		objectStorage: objectStorage,
	}
}

func (s *documentDeleteAllService) DeleteAll(ctx context.Context, ownerID int64) (int, error) {
	documents, _, err := s.repository.List(ctx, ownerID, 1000, 0)
	if err != nil {
		return 0, errors.NewPersistenceError(err)
	}

	if len(documents) == 0 {
		return 0, nil
	}

	deletedCount, err := s.repository.DeleteAllByOwnerID(ctx, ownerID)
	if err != nil {
		return 0, errors.NewPersistenceError(err)
	}

	for _, doc := range documents {
		if err := s.objectStorage.Delete(ctx, doc.ObjectKey); err != nil {
			log.Printf("failed to delete object %s from S3: %v (metadata was already deleted)", doc.ObjectKey, err)
		}
	}

	return deletedCount, nil
}
