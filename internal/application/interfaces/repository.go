package interfaces

import "github.com/kristianrpo/document-management-microservice/internal/domain"

type DocumentRepository interface {
    Create(doc *domain.Document) error
}
