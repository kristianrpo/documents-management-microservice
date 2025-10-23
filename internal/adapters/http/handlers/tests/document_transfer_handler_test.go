package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransferService struct{ mock.Mock }

func (m *mockTransferService) PrepareTransfer(ctx context.Context, ownerID int64) ([]usecases.DocumentTransferResult, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]usecases.DocumentTransferResult), args.Error(1)
}

func TestDocumentTransferHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockTransferService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentTransferHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/transfer/:id_citizen", h.PrepareTransfer)

	expiresAt := time.Now().Add(15 * time.Minute)
	transferResults := []usecases.DocumentTransferResult{
		{
			Document: &models.Document{
				ID:         "doc1",
				Filename:   "test1.pdf",
				MimeType:   "application/pdf",
				SizeBytes:  1024,
				HashSHA256: "hash1",
			},
			PresignedURL: "https://s3.example.com/doc1?token=abc",
			ExpiresAt:    expiresAt,
		},
		{
			Document: &models.Document{
				ID:         "doc2",
				Filename:   "test2.pdf",
				MimeType:   "application/pdf",
				SizeBytes:  2048,
				HashSHA256: "hash2",
			},
			PresignedURL: "https://s3.example.com/doc2?token=def",
			ExpiresAt:    expiresAt,
		},
	}

	service.On("PrepareTransfer", mock.Anything, int64(123456)).Return(transferResults, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/transfer/123456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Documents prepared for transfer")
	assert.Contains(t, w.Body.String(), `"total_documents":2`)
	assert.Contains(t, w.Body.String(), "test1.pdf")
	assert.Contains(t, w.Body.String(), "test2.pdf")
	service.AssertExpectations(t)
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentTransferHandler_ValidationError_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockTransferService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentTransferHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/transfer/:id_citizen", h.PrepareTransfer)

	// Invalid id_citizen (not a number)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/transfer/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_ID_CITIZEN")
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentTransferHandler_ValidationError_NegativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockTransferService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentTransferHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/transfer/:id_citizen", h.PrepareTransfer)

	// Negative id_citizen
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/transfer/-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_ID_CITIZEN")
}

func TestDocumentTransferHandler_PersistenceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service := new(mockTransferService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentTransferHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/transfer/:id_citizen", h.PrepareTransfer)

	service.On("PrepareTransfer", mock.Anything, int64(123456)).Return(nil, errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/transfer/123456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}

func TestDocumentTransferHandler_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockTransferService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentTransferHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/transfer/:id_citizen", h.PrepareTransfer)

	// User has no documents
	service.On("PrepareTransfer", mock.Anything, int64(999999)).Return([]usecases.DocumentTransferResult{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/transfer/999999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total_documents":0`)
	service.AssertExpectations(t)
}
