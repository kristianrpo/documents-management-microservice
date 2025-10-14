package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type ListData struct {
	Documents  []shared.DocumentResponse `json:"documents"`
	Pagination shared.Pagination         `json:"pagination"`
}

type ListResponse struct {
	Success bool     `json:"success" example:"true"`
	Data    ListData `json:"data"`
}

type ListErrorResponse struct {
	Error shared.ErrorDetail `json:"error"`
}
