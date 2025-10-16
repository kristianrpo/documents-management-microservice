package errors_test

import (
	"testing"

	httperrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectedMsg string
	}{
		{
			name:        "simple validation message",
			message:     "field is required",
			expectedMsg: "field is required",
		},
		{
			name:        "detailed validation message",
			message:     "email must be a valid email address",
			expectedMsg: "email must be a valid email address",
		},
		{
			name:        "empty message",
			message:     "",
			expectedMsg: "",
		},
		{
			name:        "message with special characters",
			message:     "value must be between 1-100",
			expectedMsg: "value must be between 1-100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &httperrors.ValidationError{Message: tt.message}
			assert.Equal(t, tt.expectedMsg, err.Error())
		})
	}
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectedMsg string
	}{
		{
			name:        "create validation error",
			message:     "invalid input",
			expectedMsg: "invalid input",
		},
		{
			name:        "create with empty message",
			message:     "",
			expectedMsg: "",
		},
		{
			name:        "create with long message",
			message:     "this is a very long validation error message that describes in detail what went wrong",
			expectedMsg: "this is a very long validation error message that describes in detail what went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httperrors.NewValidationError(tt.message)
			assert.NotNil(t, err)
			assert.Equal(t, tt.expectedMsg, err.Message)
			assert.Equal(t, tt.expectedMsg, err.Error())
		})
	}
}
