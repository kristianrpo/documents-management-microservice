package usecases

import (
	"context"
	"math"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// DocumentListService defines the interface for listing documents with pagination
type DocumentListService interface {
	List(ctx context.Context, ownerID int64, page, limit int) ([]*models.Document, util.PaginationParams, int, int64, error)
}

type documentListService struct {
	repository interfaces.DocumentRepository
}

// NewDocumentListService creates a new document list service
func NewDocumentListService(repository interfaces.DocumentRepository) DocumentListService {
	return &documentListService{
		repository: repository,
	}
}

// List retrieves a paginated list of documents for a specific owner
func (s *documentListService) List(ctx context.Context, ownerID int64, page, limit int) ([]*models.Document, util.PaginationParams, int, int64, error) {
	pagination := util.NormalizePagination(page, limit)

	documents, totalCount, err := s.repository.List(ctx, ownerID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, util.PaginationParams{}, 0, 0, errors.NewPersistenceError(err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return documents, pagination, totalPages, totalCount, nil
}
