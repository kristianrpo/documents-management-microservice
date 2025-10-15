package usecases

import (
	"context"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

type DocumentGetService interface {
	GetByID(ctx context.Context, id string) (*models.Document, error)
}

type documentGetService struct {
	repository interfaces.DocumentRepository
}

func NewDocumentGetService(repository interfaces.DocumentRepository) DocumentGetService {
	return &documentGetService{
		repository: repository,
	}
}

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
