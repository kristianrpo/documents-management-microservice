package interfaces

import "github.com/kristianrpo/document-management-microservice/internal/domain"

type DocumentRepository interface {
	Create(doc *domain.Document) error

	FindByHashAndEmail(hashSHA256, ownerEmail string) (*domain.Document, error)

	GetByID(id string) (*domain.Document, error)

	List(ownerEmail string, limit, offset int) ([]*domain.Document, int64, error)

	DeleteByID(id string) (*domain.Document, error)

	DeleteAllByEmail(email string) (int, error)
}
