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
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthGetService struct{ mock.Mock }

func (m *mockAuthGetService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

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

	getService := new(mockAuthGetService)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})
	h := handlers.NewDocumentRequestAuthenticationHandler(service, getService, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	getService.On("GetByID", mock.Anything, "doc123").Return(&models.Document{ID: "doc123", OwnerID: 123456}, nil)
	service.On("RequestAuthentication", mock.Anything, "doc123").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc123/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication request submitted successfully")
	service.AssertExpectations(t)
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentRequestAuthenticationHandler_EmptyDocumentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	getService := new(mockAuthGetService)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})
	h := handlers.NewDocumentRequestAuthenticationHandler(service, getService, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	// Empty document ID returns 400 Bad Request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents//request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_DOCUMENT_ID")
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentRequestAuthenticationHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	getService := new(mockAuthGetService)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})
	h := handlers.NewDocumentRequestAuthenticationHandler(service, getService, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	// Return not found from GetByID to simulate missing document
	getService.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.NewNotFoundError("document not found"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/nonexistent/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	service.AssertExpectations(t)
}

//nolint:dupl // similar setup across handler tests
func TestDocumentRequestAuthenticationHandler_PersistenceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockRequestAuthService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	getService := new(mockAuthGetService)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 123456})
		c.Next()
	})
	h := handlers.NewDocumentRequestAuthenticationHandler(service, getService, errHandler, metricsCollector)
	r.POST("/api/v1/documents/:id/request-authentication", h.RequestAuthentication)

	getService.On("GetByID", mock.Anything, "doc123").Return(&models.Document{ID: "doc123", OwnerID: 123456}, nil)
	service.On("RequestAuthentication", mock.Anything, "doc123").Return(errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc123/request-authentication", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
