package endpoints_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckResponse(t *testing.T) {
	response := endpoints.HealthCheckResponse{OK: true}
	assert.True(t, response.OK)
}

func TestUploadResponse(t *testing.T) {
	doc := shared.DocumentResponse{
		ID:       "doc-123",
		Filename: "test.pdf",
	}
	
	response := endpoints.UploadResponse{
		Success: true,
		Data:    doc,
	}
	
	assert.True(t, response.Success)
	assert.Equal(t, "doc-123", response.Data.ID)
	assert.Equal(t, "test.pdf", response.Data.Filename)
}

func TestUploadErrorResponse(t *testing.T) {
	err := shared.ErrorDetail{
		Code:    "VALIDATION_ERROR",
		Message: "file is required",
	}
	
	response := endpoints.UploadErrorResponse{Error: err}
	
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Equal(t, "file is required", response.Error.Message)
}

func TestListResponse(t *testing.T) {
	docs := []shared.DocumentResponse{
		{ID: "doc-1", Filename: "file1.pdf"},
		{ID: "doc-2", Filename: "file2.pdf"},
	}
	
	pagination := shared.Pagination{
		Page:       1,
		Limit:      10,
		TotalItems: 2,
		TotalPages: 1,
	}
	
	response := endpoints.ListResponse{
		Success: true,
		Data: endpoints.ListData{
			Documents:  docs,
			Pagination: pagination,
		},
	}
	
	assert.True(t, response.Success)
	assert.Len(t, response.Data.Documents, 2)
	assert.Equal(t, 1, response.Data.Pagination.Page)
}

func TestListErrorResponse(t *testing.T) {
	err := shared.ErrorDetail{
		Code:    "BAD_REQUEST",
		Message: "invalid parameters",
	}
	
	response := endpoints.ListErrorResponse{Error: err}
	
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
}

func TestGetResponse(t *testing.T) {
	doc := shared.DocumentResponse{
		ID:       "doc-456",
		Filename: "document.pdf",
		MimeType: "application/pdf",
	}
	
	response := endpoints.GetResponse{
		Success: true,
		Data:    doc,
	}
	
	assert.True(t, response.Success)
	assert.Equal(t, "doc-456", response.Data.ID)
	assert.Equal(t, "application/pdf", response.Data.MimeType)
}

func TestGetErrorResponse(t *testing.T) {
	err := shared.ErrorDetail{
		Code:    "NOT_FOUND",
		Message: "document not found",
	}
	
	response := endpoints.GetErrorResponse{Error: err}
	
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "document not found", response.Error.Message)
}

func TestDeleteResponse(t *testing.T) {
	response := endpoints.DeleteResponse{
		Success: true,
		Message: "document deleted successfully",
	}
	
	assert.True(t, response.Success)
	assert.Equal(t, "document deleted successfully", response.Message)
}

func TestDeleteErrorResponse(t *testing.T) {
	err := shared.ErrorDetail{
		Code:    "NOT_FOUND",
		Message: "document not found",
	}
	
	response := endpoints.DeleteErrorResponse{Error: err}
	
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
}

func TestTransferResponse(t *testing.T) {
	transferDocs := []shared.TransferDocument{
		{
			ID:           "doc-1",
			Filename:     "file1.pdf",
			PresignedURL: "https://example.com/file1",
			ExpiresAt:    "2025-10-16T10:00:00Z",
		},
		{
			ID:           "doc-2",
			Filename:     "file2.pdf",
			PresignedURL: "https://example.com/file2",
			ExpiresAt:    "2025-10-16T10:00:00Z",
		},
	}
	
	data := endpoints.TransferData{
		IDCitizen:      12345,
		TotalDocuments: 2,
		Documents:      transferDocs,
		ExpiresIn:      "15m",
	}
	
	response := endpoints.TransferResponse{
		Success: true,
		Message: "Documents prepared for transfer",
		Data:    data,
	}
	
	assert.True(t, response.Success)
	assert.Equal(t, "Documents prepared for transfer", response.Message)
	assert.Equal(t, int64(12345), response.Data.IDCitizen)
	assert.Equal(t, 2, response.Data.TotalDocuments)
	assert.Len(t, response.Data.Documents, 2)
	assert.Equal(t, "15m", response.Data.ExpiresIn)
}

func TestTransferErrorResponse(t *testing.T) {
	err := shared.ErrorDetail{
		Code:    "NOT_FOUND",
		Message: "no documents found",
	}
	
	response := endpoints.TransferErrorResponse{
		Success: false,
		Error:   err,
	}
	
	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
}
