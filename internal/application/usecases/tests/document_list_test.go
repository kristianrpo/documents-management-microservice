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

func TestDocumentListService_List_Success(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	service := usecases.NewDocumentListService(repo)

	ctx := context.Background()
	ownerID := int64(1)
	page := 1
	pageSize := 10
	
	docs := []*models.Document{
		{
			ID:        "doc-1",
			Filename:  "test1.pdf",
			OwnerID:   ownerID,
			SizeBytes: 1024,
			MimeType:  "application/pdf",
			CreatedAt: time.Now(),
		},
		{
			ID:        "doc-2",
			Filename:  "test2.pdf",
			OwnerID:   ownerID,
			SizeBytes: 2048,
			MimeType:  "application/pdf",
			CreatedAt: time.Now(),
		},
	}
	totalCount := int64(2)

	repo.On("List", ctx, ownerID, pageSize, 0).Return(docs, totalCount, nil)

	// Act
	result, pagination, totalPages, total, err := service.List(ctx, ownerID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 1, totalPages)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, pageSize, pagination.Limit)
	assert.Equal(t, totalCount, total)
	assert.Equal(t, "doc-1", result[0].ID)
	assert.Equal(t, "doc-2", result[1].ID)
	
	repo.AssertExpectations(t)
}

func TestDocumentListService_List_EmptyList(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	service := usecases.NewDocumentListService(repo)

	ctx := context.Background()
	ownerID := int64(1)
	page := 1
	pageSize := 10

	repo.On("List", ctx, ownerID, pageSize, 0).Return([]*models.Document{}, int64(0), nil)

	// Act
	result, pagination, totalPages, total, err := service.List(ctx, ownerID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, 1, totalPages)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, int64(0), total)
	
	repo.AssertExpectations(t)
}

func TestDocumentListService_List_SecondPage(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	service := usecases.NewDocumentListService(repo)

	ctx := context.Background()
	ownerID := int64(1)
	page := 2
	pageSize := 5
	offset := 5

	docs := []*models.Document{
		{ID: "doc-6", Filename: "test6.pdf", OwnerID: ownerID},
		{ID: "doc-7", Filename: "test7.pdf", OwnerID: ownerID},
	}
	totalCount := int64(7)

	repo.On("List", ctx, ownerID, pageSize, offset).Return(docs, totalCount, nil)

	// Act
	result, pagination, totalPages, total, err := service.List(ctx, ownerID, page, pageSize)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, pagination.Page)
	assert.Equal(t, 5, pagination.Limit)
	assert.Equal(t, 2, totalPages)
	assert.Equal(t, totalCount, total)
	
	repo.AssertExpectations(t)
}

func TestDocumentListService_List_RepositoryError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	service := usecases.NewDocumentListService(repo)

	ctx := context.Background()
	ownerID := int64(1)
	page := 1
	pageSize := 10

	expectedError := errors.New("database error")
	repo.On("List", ctx, ownerID, pageSize, 0).Return(nil, int64(0), expectedError)

	// Act
	result, pagination, totalPages, total, err := service.List(ctx, ownerID, page, pageSize)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, totalPages)
	assert.Equal(t, 0, pagination.Page)
	assert.Equal(t, int64(0), total)
	assert.Contains(t, err.Error(), "failed to persist document")
	
	repo.AssertExpectations(t)
}

