package shared_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestTransferDocument_Structure(t *testing.T) {
	t.Run("complete transfer document", func(t *testing.T) {
		doc := shared.TransferDocument{
			ID:           "123e4567-e89b-12d3-a456-426614174000",
			Filename:     "passport.pdf",
			MimeType:     "application/pdf",
			SizeBytes:    1048576,
			HashSHA256:   "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			PresignedURL: "https://s3.amazonaws.com/bucket/key?signature=abc123",
			ExpiresAt:    "2025-10-14T15:30:00Z",
		}

		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", doc.ID)
		assert.Equal(t, "passport.pdf", doc.Filename)
		assert.Equal(t, "application/pdf", doc.MimeType)
		assert.Equal(t, int64(1048576), doc.SizeBytes)
		assert.Equal(t, "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", doc.HashSHA256)
		assert.Equal(t, "https://s3.amazonaws.com/bucket/key?signature=abc123", doc.PresignedURL)
		assert.Equal(t, "2025-10-14T15:30:00Z", doc.ExpiresAt)
	})

	t.Run("transfer document with different file types", func(t *testing.T) {
		testCases := []struct {
			filename string
			mimeType string
		}{
			{"document.pdf", "application/pdf"},
			{"image.png", "image/png"},
			{"photo.jpg", "image/jpeg"},
			{"data.json", "application/json"},
			{"archive.zip", "application/zip"},
		}

		for _, tc := range testCases {
			doc := shared.TransferDocument{
				ID:           "test-id",
				Filename:     tc.filename,
				MimeType:     tc.mimeType,
				SizeBytes:    1024,
				HashSHA256:   "somehash123",
				PresignedURL: "https://example.com/presigned",
				ExpiresAt:    "2025-10-16T12:00:00Z",
			}

			assert.Equal(t, tc.filename, doc.Filename)
			assert.Equal(t, tc.mimeType, doc.MimeType)
		}
	})

	t.Run("transfer document with various sizes", func(t *testing.T) {
		sizes := []int64{1024, 1048576, 10485760, 104857600}

		for _, size := range sizes {
			doc := shared.TransferDocument{
				ID:           "test-id",
				Filename:     "test.pdf",
				MimeType:     "application/pdf",
				SizeBytes:    size,
				HashSHA256:   "hash",
				PresignedURL: "https://example.com/presigned",
				ExpiresAt:    "2025-10-16T12:00:00Z",
			}

			assert.Equal(t, size, doc.SizeBytes)
		}
	})

	t.Run("transfer document with empty fields", func(t *testing.T) {
		doc := shared.TransferDocument{}

		assert.Equal(t, "", doc.ID)
		assert.Equal(t, "", doc.Filename)
		assert.Equal(t, "", doc.MimeType)
		assert.Equal(t, int64(0), doc.SizeBytes)
		assert.Equal(t, "", doc.HashSHA256)
		assert.Equal(t, "", doc.PresignedURL)
		assert.Equal(t, "", doc.ExpiresAt)
	})
}
