package interfaces

import (
	"context"
	
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error

	FindByHashAndOwnerID(ctx context.Context, hashSHA256 string, ownerID int64) (*models.Document, error)

	GetByID(ctx context.Context, id string) (*models.Document, error)

	List(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, int64, error)

	DeleteByID(ctx context.Context, id string) (*models.Document, error)

	DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error)

	UpdateAuthenticationStatus(ctx context.Context, documentID string, status models.AuthenticationStatus) error
}
