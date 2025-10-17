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

func TestDocumentDeleteService_Execute_Success(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteService(repo, storage)

	ctx := context.Background()
	documentID := "doc-123"

	doc := &models.Document{
		ID:        documentID,
		Filename:  "test.pdf",
		OwnerID:   1,
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		ObjectKey: "documents/test.pdf",
		URL:       "https://s3.amazonaws.com/bucket/documents/test.pdf",
		CreatedAt: time.Now(),
	}

	repo.On("DeleteByID", ctx, documentID).Return(doc, nil)
	storage.On("Delete", ctx, "documents/test.pdf").Return(nil)

	// Act
	err := service.Delete(ctx, documentID)

	// Assert
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestDocumentDeleteService_Execute_DocumentNotFound(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteService(repo, storage)

	ctx := context.Background()
	documentID := "non-existent"

	repo.On("DeleteByID", ctx, documentID).Return(nil, nil)

	// Act
	err := service.Delete(ctx, documentID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	repo.AssertExpectations(t)
}

func TestDocumentDeleteService_Execute_StorageDeleteError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteService(repo, storage)

	ctx := context.Background()
	documentID := "doc-123"

	doc := &models.Document{
		ID:        documentID,
		Filename:  "test.pdf",
		OwnerID:   1,
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		ObjectKey: "documents/test.pdf",
		URL:       "https://s3.amazonaws.com/bucket/documents/test.pdf",
		CreatedAt: time.Now(),
	}

	expectedError := errors.New("storage delete failed")
	repo.On("DeleteByID", ctx, documentID).Return(doc, nil)
	storage.On("Delete", ctx, "documents/test.pdf").Return(expectedError)

	// Act
	err := service.Delete(ctx, documentID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete object from storage")

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestDocumentDeleteService_Execute_RepositoryError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)

	service := usecases.NewDocumentDeleteService(repo, storage)

	ctx := context.Background()
	documentID := "doc-123"

	expectedError := errors.New("database error")
	repo.On("DeleteByID", ctx, documentID).Return(nil, expectedError)

	// Act
	err := service.Delete(ctx, documentID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to persist document")

	repo.AssertExpectations(t)
}
