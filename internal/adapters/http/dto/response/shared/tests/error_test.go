package shared_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorDetail(t *testing.T) {
	t.Run("create error detail with code and message", func(t *testing.T) {
		errorDetail := shared.NewErrorDetail("VALIDATION_ERROR", "field is required")

		assert.Equal(t, "VALIDATION_ERROR", errorDetail.Code)
		assert.Equal(t, "field is required", errorDetail.Message)
	})

	t.Run("create error detail with empty values", func(t *testing.T) {
		errorDetail := shared.NewErrorDetail("", "")

		assert.Equal(t, "", errorDetail.Code)
		assert.Equal(t, "", errorDetail.Message)
	})

	t.Run("create error detail with different error types", func(t *testing.T) {
		testCases := []struct {
			code    string
			message string
		}{
			{"NOT_FOUND", "document not found"},
			{"INTERNAL_ERROR", "internal server error occurred"},
			{"UNAUTHORIZED", "user is not authorized"},
			{"BAD_REQUEST", "invalid request format"},
		}

		for _, tc := range testCases {
			errorDetail := shared.NewErrorDetail(tc.code, tc.message)

			assert.Equal(t, tc.code, errorDetail.Code)
			assert.Equal(t, tc.message, errorDetail.Message)
		}
	})
}

func TestErrorDetail_Structure(t *testing.T) {
	t.Run("error detail struct fields", func(t *testing.T) {
		errorDetail := shared.ErrorDetail{
			Code:    "TEST_ERROR",
			Message: "test message",
		}

		assert.Equal(t, "TEST_ERROR", errorDetail.Code)
		assert.Equal(t, "test message", errorDetail.Message)
	})
}
