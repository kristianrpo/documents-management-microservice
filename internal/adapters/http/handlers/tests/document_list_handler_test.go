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
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

// Fake services implementing DocumentListService
type okListService struct{}

func (okListService) List(ctx context.Context, ownerID int64, page, limit int) ([]*models.Document, util.PaginationParams, int, int64, error) {
	return []*models.Document{{ID: "1", Filename: "a.pdf"}}, util.PaginationParams{Page: 1, Limit: 10}, 1, 1, nil
}

type errListService struct{}

func (errListService) List(ctx context.Context, ownerID int64, page, limit int) ([]*models.Document, util.PaginationParams, int, int64, error) {
	return nil, util.PaginationParams{}, 0, 0, domainerrors.NewPersistenceError(assert.AnError)
}

func TestDocumentListHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// inject authenticated user (owner id 1)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 1})
		c.Next()
	})

	h := handlers.NewDocumentListHandler(okListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "a.pdf")
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentListHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentListHandler(okListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	// invalid id_citizen (string instead of int)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?id_citizen=abc&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

//nolint:dupl // Test setup boilerplate is similar across test files
func TestDocumentListHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	// inject authenticated user (owner id 1)
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: 1})
		c.Next()
	})

	h := handlers.NewDocumentListHandler(errListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
}
