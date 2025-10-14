package interfaces

import "github.com/kristianrpo/document-management-microservice/internal/domain"

type DocumentRepository interface {
	Create(doc *domain.Document) error

	FindByHashAndOwnerID(hashSHA256 string, ownerID int64) (*domain.Document, error)

	GetByID(id string) (*domain.Document, error)

	List(ownerID int64, limit, offset int) ([]*domain.Document, int64, error)

	DeleteByID(id string) (*domain.Document, error)

	DeleteAllByOwnerID(ownerID int64) (int, error)
}
