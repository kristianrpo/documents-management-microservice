package endpoints

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
)

// RequestAuthenticationResponse represents a successful authentication request response
type RequestAuthenticationResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Authentication request submitted successfully"`
}

// RequestAuthenticationErrorResponse represents an error response for authentication request
type RequestAuthenticationErrorResponse struct {
	Success bool               `json:"success" example:"false"`
	Error   shared.ErrorDetail `json:"error"`
}
