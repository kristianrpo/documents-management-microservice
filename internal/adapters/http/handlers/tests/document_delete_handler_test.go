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

type mockDeleteService struct{ mock.Mock }

func (m *mockDeleteService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDocumentDeleteHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentDeleteHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/:id", h.Delete)

	service.On("Delete", mock.Anything, "doc123").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/doc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "document deleted successfully")
	service.AssertExpectations(t)
}

func TestDocumentDeleteHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentDeleteHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/:id", h.Delete)

	service.On("Delete", mock.Anything, "nope").Return(errors.NewNotFoundError("document not found"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	service.AssertExpectations(t)
}

func TestDocumentDeleteHandler_PersistenceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentDeleteHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/:id", h.Delete)

	service.On("Delete", mock.Anything, "doc123").Return(errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/doc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
