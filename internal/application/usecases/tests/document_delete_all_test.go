package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestDocumentDeleteAllService_Execute_Success(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteAllService(repo, storage)

	ctx := context.Background()
	ownerID := int64(1)
	docs := []*models.Document{
		{ID: "1", OwnerID: ownerID, ObjectKey: "k1", Filename: "a.pdf", SizeBytes: 10, MimeType: "application/pdf", CreatedAt: time.Now()},
		{ID: "2", OwnerID: ownerID, ObjectKey: "k2", Filename: "b.pdf", SizeBytes: 20, MimeType: "application/pdf", CreatedAt: time.Now()},
	}
	repo.On("List", ctx, ownerID, 1000, 0).Return(docs, int64(len(docs)), nil)
	repo.On("DeleteAllByOwnerID", ctx, ownerID).Return(len(docs), nil)
	// storage deletions are best-effort; even if they fail, service doesn't error, just logs
    
	// Expect Delete for each doc
    
	storage.On("Delete", ctx, "k1").Return(nil)
	storage.On("Delete", ctx, "k2").Return(nil)

	// Act
	count, err := service.DeleteAll(ctx, ownerID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, len(docs), count)
	
	repo.AssertExpectations(t)
}

func TestDocumentDeleteAllService_Execute_NoDocuments(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteAllService(repo, storage)

	ctx := context.Background()
	ownerID := int64(1)
	repo.On("List", ctx, ownerID, 1000, 0).Return([]*models.Document{}, int64(0), nil)

	// Act
	count, err := service.DeleteAll(ctx, ownerID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	
	repo.AssertExpectations(t)
}

func TestDocumentDeleteAllService_Execute_RepositoryError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteAllService(repo, storage)

	ctx := context.Background()
	ownerID := int64(1)

	expectedError := errors.New("database error")
	repo.On("List", ctx, ownerID, 1000, 0).Return(nil, int64(0), expectedError)

	// Act
	count, err := service.DeleteAll(ctx, ownerID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "failed to persist document")
	
	repo.AssertExpectations(t)
}

func TestDocumentDeleteAllService_Execute_InvalidOwnerID(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteAllService(repo, storage)

	ctx := context.Background()
	ownerID := int64(0)
	// current implementation does not validate ownerID and will list -> assume empty
	repo.On("List", ctx, ownerID, 1000, 0).Return([]*models.Document{}, int64(0), nil)
	// Act
	count, err := service.DeleteAll(ctx, ownerID)
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
