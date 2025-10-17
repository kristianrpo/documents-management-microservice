package errors_test

import (
	"net/http"
	"testing"

	httperrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorMapper_MapDomainErrorToHTTPStatus(t *testing.T) {
	mapper := httperrors.NewErrorMapper()

	tests := []struct {
		name           string
		domainError    *domainerrors.DomainError
		expectedStatus int
	}{
		{
			name: "validation error maps to bad request",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeValidation,
				Message: "invalid input",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "file read error maps to bad request",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeFileRead,
				Message: "failed to read file",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "hash calculate error maps to internal server error",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeHashCalculate,
				Message: "failed to calculate hash",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "storage upload error maps to internal server error",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeStorageUpload,
				Message: "failed to upload to S3",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "persistence error maps to internal server error",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodePersistence,
				Message: "failed to save to database",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "not found error maps to not found",
			domainError: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeNotFound,
				Message: "document not found",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "unknown error code maps to internal server error",
			domainError: &domainerrors.DomainError{
				Code:    "UNKNOWN_ERROR",
				Message: "unknown error occurred",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "empty error code maps to internal server error",
			domainError: &domainerrors.DomainError{
				Code:    "",
				Message: "error with no code",
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mapper.MapDomainErrorToHTTPStatus(tt.domainError)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestNewErrorMapper(t *testing.T) {
	mapper := httperrors.NewErrorMapper()
	assert.NotNil(t, mapper)
}
