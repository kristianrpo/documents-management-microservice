package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type DeleteAllData struct {
	DeletedCount int `json:"deleted_count" example:"5"`
}
type DeleteAllResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"all documents deleted successfully"`
	Data    DeleteAllData `json:"data"`
}

type DeleteAllErrorResponse struct {
	Error shared.ErrorDetail `json:"error"`
}
