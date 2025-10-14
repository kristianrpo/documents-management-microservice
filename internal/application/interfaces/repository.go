package interfaces

import (
	"context"
	
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document) error

	FindByHashAndOwnerID(ctx context.Context, hashSHA256 string, ownerID int64) (*domain.Document, error)

	GetByID(ctx context.Context, id string) (*domain.Document, error)

	List(ctx context.Context, ownerID int64, limit, offset int) ([]*domain.Document, int64, error)

	DeleteByID(ctx context.Context, id string) (*domain.Document, error)

	DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error)
}
