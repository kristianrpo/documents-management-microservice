package presenter

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

func ToDocumentResponse(document *domain.Document) *shared.DocumentResponse {
	if document == nil {
		return nil
	}

	return &shared.DocumentResponse{
		ID:        document.ID,
		Filename:  document.Filename,
		MimeType:  document.MimeType,
		SizeBytes: document.SizeBytes,
		HashSHA256: document.HashSHA256,
		URL:       document.URL,
		OwnerID:   document.OwnerID,
	}
}

func ToDocumentResponseList(documents []*domain.Document) []shared.DocumentResponse {
	if documents == nil {
		return []shared.DocumentResponse{}
	}

	result := make([]shared.DocumentResponse, 0, len(documents))
	for _, doc := range documents {
		if docData := ToDocumentResponse(doc); docData != nil {
			result = append(result, *docData)
		}
	}

	return result
}
