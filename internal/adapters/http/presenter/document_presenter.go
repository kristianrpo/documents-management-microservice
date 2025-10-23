package presenter

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// ToDocumentResponse converts a domain document model to an HTTP response DTO
func ToDocumentResponse(document *models.Document) *shared.DocumentResponse {
	if document == nil {
		return nil
	}

	return &shared.DocumentResponse{
		ID:                   document.ID,
		Filename:             document.Filename,
		MimeType:             document.MimeType,
		SizeBytes:            document.SizeBytes,
		HashSHA256:           document.HashSHA256,
		URL:                  document.URL,
		OwnerID:              document.OwnerID,
		AuthenticationStatus: string(document.AuthenticationStatus),
	}
}

// ToDocumentResponseList converts a list of domain document models to a list of HTTP response DTOs
func ToDocumentResponseList(documents []*models.Document) []shared.DocumentResponse {
	if documents == nil {
		return []shared.DocumentResponse{}
	}

	result := make([]shared.DocumentResponse, 0, len(documents))
	for _, doc := range documents {
		// For list responses we intentionally omit the URL. Use a lightweight mapping.
		if listItem := toDocumentListItem(doc); listItem != nil {
			result = append(result, *listItem)
		}
	}

	return result
}

// toDocumentListItem maps a document to a response DTO used in lists (URL omitted)
func toDocumentListItem(document *models.Document) *shared.DocumentResponse {
	if document == nil {
		return nil
	}

	return &shared.DocumentResponse{
		ID:                   document.ID,
		Filename:             document.Filename,
		MimeType:             document.MimeType,
		SizeBytes:            document.SizeBytes,
		HashSHA256:           document.HashSHA256,
		// URL intentionally omitted in list responses for security/privacy
		URL:                  "",
		OwnerID:              document.OwnerID,
		AuthenticationStatus: string(document.AuthenticationStatus),
	}
}
