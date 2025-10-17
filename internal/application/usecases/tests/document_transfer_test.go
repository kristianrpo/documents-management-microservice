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

func TestNewDocumentTransferService(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	t.Run("creates service with custom expiration", func(t *testing.T) {
		expiration := 30 * time.Minute
		service := usecases.NewDocumentTransferService(mockRepo, mockStorage, expiration)
		assert.NotNil(t, service)
	})

	t.Run("creates service with default expiration when zero", func(t *testing.T) {
		service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 0)
		assert.NotNil(t, service)
	})
}

func TestPrepareTransfer_Success(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 15*time.Minute)

	ctx := context.Background()
	ownerID := int64(12345)
	documents := []*models.Document{
		{
			ID:        "doc-1",
			OwnerID:   ownerID,
			Filename:  "document1.pdf",
			ObjectKey: "documents/doc-1.pdf",
		},
		{
			ID:        "doc-2",
			OwnerID:   ownerID,
			Filename:  "document2.pdf",
			ObjectKey: "documents/doc-2.pdf",
		},
	}

	mockRepo.On("List", ctx, ownerID, 1000, 0).Return(documents, int64(2), nil)
	mockStorage.On("GeneratePresignedURL", ctx, "documents/doc-1.pdf", 15*time.Minute).
		Return("https://s3.amazonaws.com/doc-1-url", nil)
	mockStorage.On("GeneratePresignedURL", ctx, "documents/doc-2.pdf", 15*time.Minute).
		Return("https://s3.amazonaws.com/doc-2-url", nil)

	results, err := service.PrepareTransfer(ctx, ownerID)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, documents[0], results[0].Document)
	assert.Equal(t, "https://s3.amazonaws.com/doc-1-url", results[0].PresignedURL)
	assert.Equal(t, documents[1], results[1].Document)
	assert.Equal(t, "https://s3.amazonaws.com/doc-2-url", results[1].PresignedURL)
	assert.NotZero(t, results[0].ExpiresAt)
	assert.NotZero(t, results[1].ExpiresAt)
	
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestPrepareTransfer_EmptyDocumentList(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 15*time.Minute)

	ctx := context.Background()
	ownerID := int64(12345)

	mockRepo.On("List", ctx, ownerID, 1000, 0).Return([]*models.Document{}, int64(0), nil)

	results, err := service.PrepareTransfer(ctx, ownerID)

	assert.NoError(t, err)
	assert.Empty(t, results)
	mockRepo.AssertExpectations(t)
}

func TestPrepareTransfer_RepositoryListError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 15*time.Minute)

	ctx := context.Background()
	ownerID := int64(12345)
	expectedError := errors.New("database error")

	mockRepo.On("List", ctx, ownerID, 1000, 0).Return(nil, int64(0), expectedError)

	results, err := service.PrepareTransfer(ctx, ownerID)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to list documents")
	mockRepo.AssertExpectations(t)
}

func TestPrepareTransfer_PresignedURLError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 15*time.Minute)

	ctx := context.Background()
	ownerID := int64(12345)
	documents := []*models.Document{
		{
			ID:        "doc-1",
			OwnerID:   ownerID,
			Filename:  "document1.pdf",
			ObjectKey: "documents/doc-1.pdf",
		},
	}
	expectedError := errors.New("S3 error")

	mockRepo.On("List", ctx, ownerID, 1000, 0).Return(documents, int64(1), nil)
	mockStorage.On("GeneratePresignedURL", ctx, "documents/doc-1.pdf", 15*time.Minute).
		Return("", expectedError)

	results, err := service.PrepareTransfer(ctx, ownerID)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to generate pre-signed URL for document doc-1")
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestPrepareTransfer_MultipleDocumentsWithExpiration(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)

	service := usecases.NewDocumentTransferService(mockRepo, mockStorage, 30*time.Minute)

	ctx := context.Background()
	ownerID := int64(12345)
	documents := []*models.Document{
		{ID: "doc-1", OwnerID: ownerID, Filename: "doc1.pdf", ObjectKey: "documents/doc-1.pdf"},
		{ID: "doc-2", OwnerID: ownerID, Filename: "doc2.pdf", ObjectKey: "documents/doc-2.pdf"},
		{ID: "doc-3", OwnerID: ownerID, Filename: "doc3.pdf", ObjectKey: "documents/doc-3.pdf"},
	}

	mockRepo.On("List", ctx, ownerID, 1000, 0).Return(documents, int64(3), nil)
	for i := range documents {
		mockStorage.On("GeneratePresignedURL", ctx, documents[i].ObjectKey, 30*time.Minute).
			Return("https://s3.amazonaws.com/url-"+documents[i].ID, nil)
	}

	results, err := service.PrepareTransfer(ctx, ownerID)

	assert.NoError(t, err)
	assert.Len(t, results, 3)
	
	// Verify all results have the same expiration time (within 1 second tolerance)
	for i := 1; i < len(results); i++ {
		timeDiff := results[i].ExpiresAt.Sub(results[i-1].ExpiresAt)
		assert.Less(t, timeDiff.Seconds(), 1.0)
	}
	
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
