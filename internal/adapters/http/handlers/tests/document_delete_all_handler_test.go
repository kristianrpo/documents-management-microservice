package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
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
	r, errHandler, metricsCollector := newTestRouter(t, true, 123456)
	service := new(mockDeleteAllService)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/user/delete-all", h.DeleteAll)

	service.On("DeleteAll", mock.Anything, int64(123456)).Return(5, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "all documents deleted successfully")
	assert.Contains(t, w.Body.String(), `"deleted_count":5`)
	service.AssertExpectations(t)
}

func TestDocumentDeleteAllHandler_ValidationError_InvalidID(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, false, 0)
	service := new(mockDeleteAllService)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/user/delete-all", h.DeleteAll)

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestDocumentDeleteAllHandler_ValidationError_NegativeID(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, false, 0)
	service := new(mockDeleteAllService)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/user/delete-all", h.DeleteAll)

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestDocumentDeleteAllHandler_PersistenceError(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, true, 123456)
	service := new(mockDeleteAllService)
	h := handlers.NewDocumentDeleteAllHandler(service, errHandler, metricsCollector)
	r.DELETE("/api/docs/documents/user/delete-all", h.DeleteAll)

	service.On("DeleteAll", mock.Anything, int64(123456)).Return(0, errors.NewPersistenceError(assert.AnError))

	req := httptest.NewRequest(http.MethodDelete, "/api/docs/documents/user/delete-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
	service.AssertExpectations(t)
}
