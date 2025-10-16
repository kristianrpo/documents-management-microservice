package models_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   models.AuthenticationStatus
		expected bool
	}{
		{
			name:     "unauthenticated status is valid",
			status:   models.AuthenticationStatusUnauthenticated,
			expected: true,
		},
		{
			name:     "authenticating status is valid",
			status:   models.AuthenticationStatusAuthenticating,
			expected: true,
		},
		{
			name:     "authenticated status is valid",
			status:   models.AuthenticationStatusAuthenticated,
			expected: true,
		},
		{
			name:     "empty status is invalid",
			status:   "",
			expected: false,
		},
		{
			name:     "unknown status is invalid",
			status:   models.AuthenticationStatus("invalid"),
			expected: false,
		},
		{
			name:     "random string is invalid",
			status:   models.AuthenticationStatus("pending"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthenticationStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   models.AuthenticationStatus
		expected string
	}{
		{
			name:     "unauthenticated to string",
			status:   models.AuthenticationStatusUnauthenticated,
			expected: "unauthenticated",
		},
		{
			name:     "authenticating to string",
			status:   models.AuthenticationStatusAuthenticating,
			expected: "authenticating",
		},
		{
			name:     "authenticated to string",
			status:   models.AuthenticationStatusAuthenticated,
			expected: "authenticated",
		},
		{
			name:     "empty status to string",
			status:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
