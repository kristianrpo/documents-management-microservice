package endpoints

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
)

// TransferResponse represents a successful transfer response
type TransferResponse struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"Documents prepared for transfer"`
	Data    TransferData `json:"data"`
}

// TransferData contains the list of documents with pre-signed URLs
type TransferData struct {
	IDCitizen      int64                     `json:"id_citizen" example:"123456789"`
	TotalDocuments int                       `json:"total_documents" example:"5"`
	Documents      []shared.TransferDocument `json:"documents"`
	ExpiresIn      string                    `json:"expires_in" example:"15m"`
}

// TransferErrorResponse represents an error response for transfer endpoint
type TransferErrorResponse struct {
	Success bool               `json:"success" example:"false"`
	Error   shared.ErrorDetail `json:"error"`
}
