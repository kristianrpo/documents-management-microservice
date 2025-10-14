package presenter

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

func ToDocumentResponse(document *domain.Document) *response.DocumentResponse {
	if document == nil {
		return nil
	}
	
	return &response.DocumentResponse{
		ID:         document.ID,
		Filename:   document.Filename,
		MimeType:   document.MimeType,
		SizeBytes:  document.SizeBytes,
		HashSHA256: document.HashSHA256,
		URL:        document.URL,
		OwnerEmail: document.OwnerEmail,
	}
}
