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
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDeleteService struct{ mock.Mock }

func (m *mockDeleteService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockDeleteGetService struct{ mock.Mock }

func (m *mockDeleteGetService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

//nolint:dupl // test boilerplate duplicated across handlers (intentional)
func TestDocumentDeleteHandler_Success(t *testing.T) {
	service := new(mockDeleteService)
	getService := new(mockDeleteGetService)

	// Expect GetByID called to verify owner
	getService.On("GetByID", mock.Anything, "doc123").Return(&models.Document{ID: "doc123", OwnerID: 123456}, nil)
	service.On("Delete", mock.Anything, "doc123").Return(nil)

	w := runWithAuthenticatedRouter(t, http.MethodDelete, "/api/docs/documents/doc123", func(r *gin.Engine) {
		errMapper := apierrors.NewErrorMapper()
		errHandler := apierrors.NewErrorHandler(errMapper)
		metricsCollector := createTestMetrics(t)
		h := handlers.NewDocumentDeleteHandler(service, getService, errHandler, metricsCollector)
		r.DELETE("/api/docs/documents/:id", h.Delete)
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "document deleted successfully")
	service.AssertExpectations(t)
}

func TestDocumentDeleteHandler_NotFound(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, true, 123456)
	service := new(mockDeleteService)
	getService := new(mockDeleteGetService)
	h := handlers.NewDocumentDeleteHandler(service, getService, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/:id", h.Delete)

	// GetByID returns a document; Delete returns not found for this id
	getService.On("GetByID", mock.Anything, "nope").Return(&models.Document{ID: "nope", OwnerID: 123456}, nil)
	service.On("Delete", mock.Anything, "nope").Return(errors.NewNotFoundError("document not found"))

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
	service.AssertExpectations(t)
}

func TestDocumentDeleteHandler_PersistenceError(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, true, 123456)
	service := new(mockDeleteService)
	getService := new(mockDeleteGetService)
	h := handlers.NewDocumentDeleteHandler(service, getService, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/:id", h.Delete)

	getService.On("GetByID", mock.Anything, "doc123").Return(&models.Document{ID: "doc123", OwnerID: 123456}, nil)
	service.On("Delete", mock.Anything, "doc123").Return(errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/doc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
