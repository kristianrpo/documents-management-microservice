package handlers_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
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

func TestDocumentUploadHandler_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		withAuth        bool
		ownerID         int64
		service         usecases.DocumentService
		buildBody       func() (*bytes.Buffer, string)
		expectedStatus  int
		expectedContent string
	}{
		{
			name:     "success",
			withAuth: true,
			ownerID:  1,
			service:  okUploadService{},
			buildBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(t, "a.pdf", "content", "1")
			},
			expectedStatus:  http.StatusCreated,
			expectedContent: "a.pdf",
		},
		{
			name:     "validation error",
			withAuth: false,
			ownerID:  0,
			service:  okUploadService{},
			buildBody: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				w := multipart.NewWriter(body)
				_ = w.Close()
				return body, w.FormDataContentType()
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContent: "VALIDATION_ERROR",
		},
		{
			name:     "service error",
			withAuth: true,
			ownerID:  1,
			service:  errUploadService{},
			buildBody: func() (*bytes.Buffer, string) {
				return createMultipartBody(t, "a.pdf", "content", "1")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContent: "PERSISTENCE_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, errHandler, metricsCollector := newTestRouter(t, tc.withAuth, tc.ownerID)
			h := handlers.NewDocumentUploadHandler(tc.service, errHandler, metricsCollector)
			r.POST("/api/docs/documents", h.Upload)

			body, contentType := tc.buildBody()
			req := httptest.NewRequest(http.MethodPost, "/api/docs/documents", body)
			req.Header.Set("Content-Type", contentType)
			wr := httptest.NewRecorder()
			r.ServeHTTP(wr, req)

			assert.Equal(t, tc.expectedStatus, wr.Code)
			assert.Contains(t, wr.Body.String(), tc.expectedContent)
		})
	}
}

// createMultipartBody builds a multipart/form-data body with a single file and id_citizen field.
func createMultipartBody(t *testing.T, filename, content, idCitizen string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", filename)
	_, _ = fw.Write([]byte(content))
	_ = w.WriteField("id_citizen", idCitizen)
	_ = w.Close()
	return body, w.FormDataContentType()
}
