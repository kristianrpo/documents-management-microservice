package handlers_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

// Fake upload service implementing DocumentService
type okUploadService struct{}

func (okUploadService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerID int64) (*models.Document, error) {
	return &models.Document{ID: "1", Filename: fileHeader.Filename, OwnerID: ownerID, MimeType: "application/pdf"}, nil
}

type errUploadService struct{}

func (errUploadService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerID int64) (*models.Document, error) {
	return nil, errors.NewPersistenceError(assert.AnError)
}

func TestDocumentUploadHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentUploadHandler(okUploadService{}, errHandler, metricsCollector)
	r.POST("/api/v1/documents", h.Upload)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "a.pdf")
	_, _ = fw.Write([]byte("content"))
	_ = w.WriteField("id_citizen", "1")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	wr := httptest.NewRecorder()
	r.ServeHTTP(wr, req)

	assert.Equal(t, http.StatusCreated, wr.Code)
	assert.Contains(t, wr.Body.String(), "a.pdf")
}

func TestDocumentUploadHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentUploadHandler(okUploadService{}, errHandler, metricsCollector)
	r.POST("/api/v1/documents", h.Upload)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	// missing file and/or id_citizen
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	wr := httptest.NewRecorder()
	r.ServeHTTP(wr, req)

	assert.Equal(t, http.StatusBadRequest, wr.Code)
	assert.Contains(t, wr.Body.String(), "VALIDATION_ERROR")
}

func TestDocumentUploadHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	h := handlers.NewDocumentUploadHandler(errUploadService{}, errHandler, metricsCollector)
	r.POST("/api/v1/documents", h.Upload)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "a.pdf")
	_, _ = fw.Write([]byte("content"))
	_ = w.WriteField("id_citizen", "1")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	wr := httptest.NewRecorder()
	r.ServeHTTP(wr, req)

	assert.Equal(t, http.StatusInternalServerError, wr.Code)
	assert.Contains(t, wr.Body.String(), "PERSISTENCE_ERROR")
}
