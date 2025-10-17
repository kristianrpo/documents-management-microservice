package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	domainErrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDocumentRequestAuthenticationService(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	t.Run("creates service with custom expiration", func(t *testing.T) {
		expiration := 12 * time.Hour
		service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"test-queue",
		expiration,
	)
		assert.NotNil(t, service)
	})

	t.Run("creates service with default expiration when zero", func(t *testing.T) {
		service := usecases.NewDocumentRequestAuthenticationService(
			mockRepo,
			mockStorage,
			mockPublisher,
			"test-queue",
			0,
		)
		assert.NotNil(t, service)
	})
}

func TestRequestAuthentication_Success(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "doc-123"
	document := &models.Document{
		ID:       documentID,
		OwnerID:  12345,
		Filename: "test-document.pdf",
		ObjectKey: "documents/test-document.pdf",
	}
	presignedURL := "https://s3.amazonaws.com/presigned-url"

	mockRepo.On("GetByID", ctx, documentID).Return(document, nil)
	mockRepo.On("UpdateAuthenticationStatus", ctx, documentID, models.AuthenticationStatusAuthenticating).Return(nil)
	mockStorage.On("GeneratePresignedURL", ctx, document.ObjectKey, 24*time.Hour).Return(presignedURL, nil)
	mockPublisher.On("Publish", ctx, "auth-queue", mock.AnythingOfType("[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		// Verify the event was marshalled correctly
		eventJSON := args.Get(2).([]byte)
		var event events.DocumentAuthenticationRequestedEvent
		err := json.Unmarshal(eventJSON, &event)
		assert.NoError(t, err)
		assert.Equal(t, document.OwnerID, event.IDCitizen)
		assert.Equal(t, presignedURL, event.URLDocument)
		assert.Equal(t, document.Filename, event.DocumentTitle)
		assert.Equal(t, documentID, event.DocumentID)
	})

	err := service.RequestAuthentication(ctx, documentID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestRequestAuthentication_DocumentNotFound(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "non-existent"

	mockRepo.On("GetByID", ctx, documentID).Return(nil, nil)

	err := service.RequestAuthentication(ctx, documentID)

	// Expect a DomainError with NOT_FOUND code
	assert.Error(t, err)
	var derr *domainErrors.DomainError
	if assert.True(t, errors.As(err, &derr)) {
		assert.Equal(t, domainErrors.ErrCodeNotFound, derr.Code)
	}
	mockRepo.AssertExpectations(t)
}

func TestRequestAuthentication_RepositoryGetError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "doc-123"
	expectedError := errors.New("database connection error")

	mockRepo.On("GetByID", ctx, documentID).Return(nil, expectedError)

	err := service.RequestAuthentication(ctx, documentID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRequestAuthentication_UpdateStatusError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "doc-123"
	document := &models.Document{
		ID:        documentID,
		OwnerID:   12345,
		Filename:  "test.pdf",
		ObjectKey: "documents/test.pdf",
	}
	expectedError := errors.New("update failed")

	mockRepo.On("GetByID", ctx, documentID).Return(document, nil)
	mockRepo.On("UpdateAuthenticationStatus", ctx, documentID, models.AuthenticationStatusAuthenticating).Return(expectedError)

	err := service.RequestAuthentication(ctx, documentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update authentication status")
	mockRepo.AssertExpectations(t)
}

func TestRequestAuthentication_PresignedURLError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "doc-123"
	document := &models.Document{
		ID:        documentID,
		OwnerID:   12345,
		Filename:  "test.pdf",
		ObjectKey: "documents/test.pdf",
	}
	expectedError := errors.New("S3 error")

	mockRepo.On("GetByID", ctx, documentID).Return(document, nil)
	mockRepo.On("UpdateAuthenticationStatus", ctx, documentID, models.AuthenticationStatusAuthenticating).Return(nil)
	mockStorage.On("GeneratePresignedURL", ctx, document.ObjectKey, 24*time.Hour).Return("", expectedError)

	err := service.RequestAuthentication(ctx, documentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate pre-signed URL")
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestRequestAuthentication_PublishError(t *testing.T) {
	mockRepo := new(MockDocumentRepository)
	mockStorage := new(MockObjectStorage)
	mockPublisher := new(MockMessagePublisher)

	service := usecases.NewDocumentRequestAuthenticationService(
		mockRepo,
		mockStorage,
		mockPublisher,
		"auth-queue",
		24*time.Hour,
	)

	ctx := context.Background()
	documentID := "doc-123"
	document := &models.Document{
		ID:        documentID,
		OwnerID:   12345,
		Filename:  "test.pdf",
		ObjectKey: "documents/test.pdf",
	}
	presignedURL := "https://s3.amazonaws.com/presigned-url"
	expectedError := errors.New("RabbitMQ connection error")

	mockRepo.On("GetByID", ctx, documentID).Return(document, nil)
	mockRepo.On("UpdateAuthenticationStatus", ctx, documentID, models.AuthenticationStatusAuthenticating).Return(nil)
	mockStorage.On("GeneratePresignedURL", ctx, document.ObjectKey, 24*time.Hour).Return(presignedURL, nil)
	mockPublisher.On("Publish", ctx, "auth-queue", mock.AnythingOfType("[]uint8")).Return(expectedError)

	err := service.RequestAuthentication(ctx, documentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish authentication request event")
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
