package interfaces

import (
	"context"

	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentRepository defines the interface for document persistence operations
type DocumentRepository interface {
	// Create stores a new document in the repository
	Create(ctx context.Context, doc *models.Document) error

	// FindByHashAndOwnerID retrieves a document by its hash and owner ID (for deduplication)
	FindByHashAndOwnerID(ctx context.Context, hashSHA256 string, ownerID int64) (*models.Document, error)

	// GetByID retrieves a document by its unique identifier
	GetByID(ctx context.Context, id string) (*models.Document, error)

	// List retrieves a paginated list of documents for a specific owner
	List(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, int64, error)

	// DeleteByID removes a document by its ID and returns the deleted document
	DeleteByID(ctx context.Context, id string) (*models.Document, error)

	// DeleteAllByOwnerID removes all documents owned by a specific user
	DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error)

	// UpdateAuthenticationStatus updates the authentication status of a document
	UpdateAuthenticationStatus(ctx context.Context, documentID string, status models.AuthenticationStatus) error
	
	// EnsureTableExists ensures the documents table exists (implementation-specific)
	// Called automatically on initialization
	EnsureTableExists(ctx context.Context) error
}
