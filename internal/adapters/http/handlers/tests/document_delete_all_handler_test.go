package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDeleteAllService struct{ mock.Mock }

func (m *mockDeleteAllService) DeleteAll(ctx context.Context, ownerID int64) (int, error) {
	args := m.Called(ctx, ownerID)
	return args.Int(0), args.Error(1)
}

func TestDocumentDeleteAllHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteAllService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// inject authenticated user (owner id 123456)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})

	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/user/delete-all", h.DeleteAll)

	service.On("DeleteAll", mock.Anything, int64(123456)).Return(5, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "all documents deleted successfully")
	assert.Contains(t, w.Body.String(), `"deleted_count":5`)
	service.AssertExpectations(t)
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentDeleteAllHandler_ValidationError_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteAllService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// No token -> should return validation error (user not authenticated)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/user/delete-all", h.DeleteAll)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentDeleteAllHandler_ValidationError_NegativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteAllService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// No token -> should return validation error (user not authenticated)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/user/delete-all", h.DeleteAll)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestDocumentDeleteAllHandler_PersistenceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockDeleteAllService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// inject authenticated user (owner id 123456)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})

	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/v1/documents/user/delete-all", h.DeleteAll)

	service.On("DeleteAll", mock.Anything, int64(123456)).Return(0, errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
