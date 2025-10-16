package shared_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestDocumentResponse_Structure(t *testing.T) {
	t.Run("complete document response", func(t *testing.T) {
		doc := shared.DocumentResponse{
			ID:                   "doc-123",
			Filename:             "test.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            1024,
			HashSHA256:           "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			URL:                  "https://example.com/doc.pdf",
			OwnerID:              12345,
			AuthenticationStatus: "unauthenticated",
		}

		assert.Equal(t, "doc-123", doc.ID)
		assert.Equal(t, "test.pdf", doc.Filename)
		assert.Equal(t, "application/pdf", doc.MimeType)
		assert.Equal(t, int64(1024), doc.SizeBytes)
		assert.Equal(t, "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", doc.HashSHA256)
		assert.Equal(t, "https://example.com/doc.pdf", doc.URL)
		assert.Equal(t, int64(12345), doc.OwnerID)
		assert.Equal(t, "unauthenticated", doc.AuthenticationStatus)
	})

	t.Run("authenticated document", func(t *testing.T) {
		doc := shared.DocumentResponse{
			ID:                   "doc-456",
			Filename:             "verified.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            2048,
			HashSHA256:           "b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3",
			URL:                  "https://example.com/verified.pdf",
			OwnerID:              67890,
			AuthenticationStatus: "authenticated",
		}

		assert.Equal(t, "authenticated", doc.AuthenticationStatus)
	})

	t.Run("authenticating document", func(t *testing.T) {
		doc := shared.DocumentResponse{
			ID:                   "doc-789",
			Filename:             "pending.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            512,
			HashSHA256:           "c5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4",
			URL:                  "https://example.com/pending.pdf",
			OwnerID:              11111,
			AuthenticationStatus: "authenticating",
		}

		assert.Equal(t, "authenticating", doc.AuthenticationStatus)
	})

	t.Run("document with various mime types", func(t *testing.T) {
		mimeTypes := []string{
			"application/pdf",
			"image/png",
			"image/jpeg",
			"application/json",
			"text/plain",
		}

		for _, mimeType := range mimeTypes {
			doc := shared.DocumentResponse{
				ID:       "doc-test",
				MimeType: mimeType,
			}

			assert.Equal(t, mimeType, doc.MimeType)
		}
	})
}
