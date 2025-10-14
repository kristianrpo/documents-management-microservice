package presenter

import (
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

func ToDocumentData(document *domain.Document) *endpoints.DocumentData {
	if document == nil {
		return nil
	}

	return &endpoints.DocumentData{
		ID:         document.ID,
		Filename:   document.Filename,
		MimeType:   document.MimeType,
		SizeBytes:  document.SizeBytes,
		HashSHA256: document.HashSHA256,
		URL:        document.URL,
		OwnerEmail: document.OwnerEmail,
	}
}

func ToDocumentDataList(documents []*domain.Document) []endpoints.DocumentData {
	if documents == nil {
		return []endpoints.DocumentData{}
	}

	result := make([]endpoints.DocumentData, 0, len(documents))
	for _, doc := range documents {
		if docData := ToDocumentData(doc); docData != nil {
			result = append(result, *docData)
		}
	}

	return result
}

