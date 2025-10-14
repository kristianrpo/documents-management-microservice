package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type GetResponse struct {
	Success bool                    `json:"success" example:"true"`
	Data    shared.DocumentResponse `json:"data"`
}

type GetErrorResponse struct {
	Error shared.ErrorDetail `json:"error"`
}
