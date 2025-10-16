package shared_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	t.Run("success response with string data", func(t *testing.T) {
		response := shared.NewSuccessResponse("test data")

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.Equal(t, "test data", *response.Data)
		assert.Nil(t, response.Error)
	})

	t.Run("success response with struct data", func(t *testing.T) {
		type TestData struct {
			ID   int
			Name string
		}
		data := TestData{ID: 1, Name: "test"}
		response := shared.NewSuccessResponse(data)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.Equal(t, 1, response.Data.ID)
		assert.Equal(t, "test", response.Data.Name)
		assert.Nil(t, response.Error)
	})

	t.Run("success response with slice data", func(t *testing.T) {
		data := []int{1, 2, 3}
		response := shared.NewSuccessResponse(data)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.Len(t, *response.Data, 3)
		assert.Nil(t, response.Error)
	})

	t.Run("success response with map data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		response := shared.NewSuccessResponse(data)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.Equal(t, "value", (*response.Data)["key"])
		assert.Nil(t, response.Error)
	})

	t.Run("success response with nil-able data", func(t *testing.T) {
		var data *string
		response := shared.NewSuccessResponse(data)

		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.Nil(t, response.Error)
	})
}

func TestNewErrorResponse(t *testing.T) {
	t.Run("error response with code and message", func(t *testing.T) {
		response := shared.NewErrorResponse("VALIDATION_ERROR", "invalid input")

		assert.False(t, response.Success)
		assert.Nil(t, response.Data)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
		assert.Equal(t, "invalid input", response.Error.Message)
	})

	t.Run("error response with empty strings", func(t *testing.T) {
		response := shared.NewErrorResponse("", "")

		assert.False(t, response.Success)
		assert.Nil(t, response.Data)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "", response.Error.Code)
		assert.Equal(t, "", response.Error.Message)
	})

	t.Run("error response with different error codes", func(t *testing.T) {
		codes := []string{"NOT_FOUND", "INTERNAL_ERROR", "UNAUTHORIZED"}
		messages := []string{"resource not found", "internal server error", "unauthorized access"}

		for i, code := range codes {
			response := shared.NewErrorResponse(code, messages[i])

			assert.False(t, response.Success)
			assert.Equal(t, code, response.Error.Code)
			assert.Equal(t, messages[i], response.Error.Message)
		}
	})
}
