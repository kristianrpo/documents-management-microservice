package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRequestAuthService struct{ mock.Mock }

func (m *mockRequestAuthService) RequestAuthentication(ctx context.Context, documentID string) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

func TestDocumentRequestAuthenticationHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentRequestAuthenticationHandler(service, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	service.On("RequestAuthentication", mock.Anything, "doc123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc123/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication request submitted successfully")
	service.AssertExpectations(t)
}

func TestDocumentRequestAuthenticationHandler_EmptyDocumentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentRequestAuthenticationHandler(service, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	// Empty document ID returns 400 Bad Request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents//request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_DOCUMENT_ID")
}

func TestDocumentRequestAuthenticationHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentRequestAuthenticationHandler(service, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	service.On("RequestAuthentication", mock.Anything, "nonexistent").Return(errors.NewNotFoundError("document not found"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/nonexistent/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	service.AssertExpectations(t)
}

func TestDocumentRequestAuthenticationHandler_PersistenceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentRequestAuthenticationHandler(service, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	service.On("RequestAuthentication", mock.Anything, "doc123").Return(errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc123/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
