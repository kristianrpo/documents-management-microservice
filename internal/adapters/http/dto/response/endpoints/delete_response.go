package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type DeleteResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"document deleted successfully"`
}

type DeleteErrorResponse struct {
	Error shared.ErrorDetail `json:"error"`
}
