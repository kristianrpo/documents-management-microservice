package repository

import (
	"gorm.io/gorm"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type documentRepository struct {
	database *gorm.DB
}

func NewDocumentRepo(database *gorm.DB) interfaces.DocumentRepository {
	return &documentRepository{database: database}
}

func (repo *documentRepository) Create(document *domain.Document) error {
	return repo.database.Create(document).Error
}
