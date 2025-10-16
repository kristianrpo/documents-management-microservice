package errors_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	httperrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorHandler(t *testing.T) {
	mapper := httperrors.NewErrorMapper()
	handler := httperrors.NewErrorHandler(mapper)

	assert.NotNil(t, handler)
}

func TestErrorHandler_HandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		error              error
		expectedStatus     int
		expectedCode       string
		expectedMessage    string
		debugMode          bool
		shouldHaveDetails  bool
	}{
		{
			name:             "validation error",
			error:            httperrors.NewValidationError("invalid input"),
			expectedStatus:   http.StatusBadRequest,
			expectedCode:     "VALIDATION_ERROR",
			expectedMessage:  "invalid input",
			debugMode:        false,
			shouldHaveDetails: false,
		},
		{
			name:             "domain validation error",
			error:            domainerrors.NewValidationError("field is required"),
			expectedStatus:   http.StatusBadRequest,
			expectedCode:     domainerrors.ErrCodeValidation,
			expectedMessage:  "field is required",
			debugMode:        false,
			shouldHaveDetails: false,
		},
		{
			name:             "domain not found error",
			error:            domainerrors.NewNotFoundError("document not found"),
			expectedStatus:   http.StatusNotFound,
			expectedCode:     domainerrors.ErrCodeNotFound,
			expectedMessage:  "document not found",
			debugMode:        false,
			shouldHaveDetails: false,
		},
		{
			name:             "domain storage error",
			error:            domainerrors.NewStorageUploadError(errors.New("S3 failed")),
			expectedStatus:   http.StatusInternalServerError,
			expectedCode:     domainerrors.ErrCodeStorageUpload,
			expectedMessage:  "failed to upload to storage",
			debugMode:        false,
			shouldHaveDetails: false,
		},
		{
			name:             "domain storage error with debug",
			error:            domainerrors.NewStorageUploadError(errors.New("S3 connection timeout")),
			expectedStatus:   http.StatusInternalServerError,
			expectedCode:     domainerrors.ErrCodeStorageUpload,
			expectedMessage:  "failed to upload to storage",
			debugMode:        true,
			shouldHaveDetails: true,
		},
		{
			name:             "generic error",
			error:            errors.New("unexpected error"),
			expectedStatus:   http.StatusInternalServerError,
			expectedCode:     "INTERNAL_ERROR",
			expectedMessage:  "an unexpected error occurred",
			debugMode:        false,
			shouldHaveDetails: false,
		},
		{
			name:             "generic error with debug",
			error:            errors.New("database connection failed"),
			expectedStatus:   http.StatusInternalServerError,
			expectedCode:     "INTERNAL_ERROR",
			expectedMessage:  "an unexpected error occurred",
			debugMode:        true,
			shouldHaveDetails: true,
		},
		{
			name:             "nil error",
			error:            nil,
			expectedStatus:   http.StatusInternalServerError,
			expectedCode:     "INTERNAL_ERROR",
			expectedMessage:  "an unexpected error occurred",
			debugMode:        false,
			shouldHaveDetails: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set debug mode
			if tt.debugMode {
				os.Setenv("DEBUG", "true")
			} else {
				os.Unsetenv("DEBUG")
			}

			mapper := httperrors.NewErrorMapper()
			handler := httperrors.NewErrorHandler(mapper)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			handler.HandleError(ctx, tt.error)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// Parse response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			assert.False(t, response["success"].(bool))
			errorData := response["error"].(map[string]interface{})
			assert.Equal(t, tt.expectedCode, errorData["code"])
			assert.Equal(t, tt.expectedMessage, errorData["message"])

			if tt.shouldHaveDetails {
				assert.Contains(t, errorData, "details")
			} else {
				assert.NotContains(t, errorData, "details")
			}
		})
	}
}
