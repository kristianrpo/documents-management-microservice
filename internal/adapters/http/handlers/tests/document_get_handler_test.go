package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGetService struct{ mock.Mock }

func (m *mockGetService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func TestDocumentGetHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockGetService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentGetHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/:id", h.GetByID)

	doc := &models.Document{ID: "123", Filename: "a.pdf", MimeType: "application/pdf"}
	service.On("GetByID", mock.Anything, "123").Return(doc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	presented := presenter.ToDocumentResponse(doc)
	assert.Contains(t, w.Body.String(), presented.Filename)
	service.AssertExpectations(t)
}

func TestDocumentGetHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockGetService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentGetHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/:id", h.GetByID)

	service.On("GetByID", mock.Anything, "nope").Return(nil, errors.NewNotFoundError("document not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}

func TestDocumentGetHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := new(mockGetService)
	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentGetHandler(service, errHandler, metricsCollector)
	r.GET("/api/v1/documents/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code) // route not matched (missing :id)
}
