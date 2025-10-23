package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDocumentGetService_GetByID_Success(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	service := usecases.NewDocumentGetService(repo, mockStorage)

	ctx := context.Background()
	documentID := "doc-123"

	doc := &models.Document{
		ID:        documentID,
		Filename:  "test.pdf",
		OwnerID:   1,
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		URL:       "https://s3.amazonaws.com/bucket/doc.pdf",
		CreatedAt: time.Now(),
	}

	repo.On("GetByID", ctx, documentID).Return(doc, nil)
	// Expect a presigned URL to be generated for the object's key
	mockStorage.On("GeneratePresignedURL", ctx, doc.ObjectKey, mock.Anything).Return("https://presigned.example.com/doc.pdf", nil)

	// Act
	result, err := service.GetByID(ctx, documentID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, documentID, result.ID)

	repo.AssertExpectations(t)
}

func TestDocumentGetService_GetByID_DocumentNotFound(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	service := usecases.NewDocumentGetService(repo, mockStorage)

	ctx := context.Background()
	documentID := "non-existent"

	repo.On("GetByID", ctx, documentID).Return(nil, nil)

	// Act
	result, err := service.GetByID(ctx, documentID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}

func TestDocumentGetService_GetByID_RepositoryError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	service := usecases.NewDocumentGetService(repo, mockStorage)

	ctx := context.Background()
	documentID := "doc-123"

	expectedError := errors.New("db error")
	repo.On("GetByID", ctx, documentID).Return(nil, expectedError)

	// Act
	result, err := service.GetByID(ctx, documentID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to persist document")

	repo.AssertExpectations(t)
}
