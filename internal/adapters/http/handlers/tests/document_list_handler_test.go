package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
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
	r, errHandler, metricsCollector := newTestRouter(t, true, 1)
	h := handlers.NewDocumentListHandler(okListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "a.pdf")
}

func TestDocumentListHandler_ValidationError(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, false, 0)
	h := handlers.NewDocumentListHandler(okListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	// invalid id_citizen (string instead of int)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?id_citizen=abc&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestDocumentListHandler_ServiceError(t *testing.T) {
	r, errHandler, metricsCollector := newTestRouter(t, true, 1)
	h := handlers.NewDocumentListHandler(errListService{}, errHandler, metricsCollector)
	r.GET("/api/v1/documents", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "PERSISTENCE_ERROR")
}
