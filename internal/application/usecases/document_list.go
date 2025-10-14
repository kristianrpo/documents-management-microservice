package usecases

import (
	"context"
	"math"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type DocumentListService interface {
	List(ctx context.Context, ownerID int64, page, limit int) ([]*domain.Document, util.PaginationParams, int, int64, error)
}

type documentListService struct {
	repository interfaces.DocumentRepository
}

func NewDocumentListService(repository interfaces.DocumentRepository) DocumentListService {
	return &documentListService{
		repository: repository,
	}
}

func (s *documentListService) List(ctx context.Context, ownerID int64, page, limit int) ([]*domain.Document, util.PaginationParams, int, int64, error) {
	pagination := util.NormalizePagination(page, limit)

	documents, totalCount, err := s.repository.List(ownerID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, util.PaginationParams{}, 0, 0, domain.NewPersistenceError(err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return documents, pagination, totalPages, totalCount, nil
}
