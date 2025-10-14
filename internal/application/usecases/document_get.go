package usecases

import (
	"context"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type DocumentGetService interface {
	GetByID(ctx context.Context, id string) (*domain.Document, error)
}

type documentGetService struct {
	repository interfaces.DocumentRepository
}

func NewDocumentGetService(repository interfaces.DocumentRepository) DocumentGetService {
	return &documentGetService{
		repository: repository,
	}
}

func (s *documentGetService) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	document, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewPersistenceError(err)
	}

	if document == nil {
		return nil, domain.NewNotFoundError("document not found")
	}

	return document, nil
}
