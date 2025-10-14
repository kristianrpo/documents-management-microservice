package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type UploadResponse struct {
	Success bool                    `json:"success" example:"true"`
	Data    shared.DocumentResponse `json:"data"`
}

type UploadErrorResponse struct {
	Error shared.ErrorDetail `json:"error"`
}
