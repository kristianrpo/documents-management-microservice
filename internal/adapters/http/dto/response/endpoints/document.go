package endpoints

import "github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"

type DocumentData struct {
	ID         string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Filename   string `json:"filename" example:"my-document.pdf"`
	MimeType   string `json:"mime_type" example:"application/pdf"`
	SizeBytes  int64  `json:"size_bytes" example:"102400"`
	HashSHA256 string `json:"hash_sha256" example:"abc123def456789..."`
	URL        string `json:"url" example:"https://my-bucket.s3.amazonaws.com/ab/abc123def456.pdf"`
	OwnerEmail string `json:"owner_email" example:"user@example.com"`
}

type DocumentUploadSuccessResponse struct {
	Success bool         `json:"success" example:"true"`
	Data    DocumentData `json:"data"`
}

type DocumentUploadErrorResponse struct {
	Success bool                `json:"success" example:"false"`
	Error   shared.ErrorDetail `json:"error"`
}

type PaginationMetadata struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	TotalItems int64 `json:"total_items" example:"50"`
	TotalPages int   `json:"total_pages" example:"5"`
}

type DocumentListData struct {
	Documents  []DocumentData     `json:"documents"`
	Pagination PaginationMetadata `json:"pagination"`
}

type DocumentListSuccessResponse struct {
	Success bool             `json:"success" example:"true"`
	Data    DocumentListData `json:"data"`
}

type DocumentListErrorResponse struct {
	Success bool                `json:"success" example:"false"`
	Error   shared.ErrorDetail `json:"error"`
}
