package presenter_test

import (
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestToDocumentResponse(t *testing.T) {
	t.Run("valid document", func(t *testing.T) {
		now := time.Now()
		doc := &models.Document{
			ID:                   "doc-123",
			Filename:             "test.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            1024,
			HashSHA256:           "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			Bucket:               "test-bucket",
			ObjectKey:            "a3/a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2.pdf",
			URL:                  "https://example.com/doc.pdf",
			OwnerID:              12345,
			AuthenticationStatus: models.AuthenticationStatusUnauthenticated,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		response := presenter.ToDocumentResponse(doc)

		assert.NotNil(t, response)
		assert.Equal(t, "doc-123", response.ID)
		assert.Equal(t, "test.pdf", response.Filename)
		assert.Equal(t, "application/pdf", response.MimeType)
		assert.Equal(t, int64(1024), response.SizeBytes)
		assert.Equal(t, "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", response.HashSHA256)
		assert.Equal(t, "https://example.com/doc.pdf", response.URL)
		assert.Equal(t, int64(12345), response.OwnerID)
		assert.Equal(t, "unauthenticated", response.AuthenticationStatus)
	})

	t.Run("nil document", func(t *testing.T) {
		response := presenter.ToDocumentResponse(nil)
		assert.Nil(t, response)
	})

	t.Run("document with authenticated status", func(t *testing.T) {
		doc := &models.Document{
			ID:                   "doc-456",
			Filename:             "verified.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            2048,
			HashSHA256:           "b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0c9d8e7f6a5b4c3",
			Bucket:               "verified-bucket",
			ObjectKey:            "b4/hash.pdf",
			URL:                  "https://example.com/verified.pdf",
			OwnerID:              67890,
			AuthenticationStatus: models.AuthenticationStatusAuthenticated,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		response := presenter.ToDocumentResponse(doc)

		assert.NotNil(t, response)
		assert.Equal(t, "authenticated", response.AuthenticationStatus)
	})

	t.Run("document with authenticating status", func(t *testing.T) {
		doc := &models.Document{
			ID:                   "doc-789",
			Filename:             "pending.pdf",
			MimeType:             "application/pdf",
			SizeBytes:            512,
			HashSHA256:           "c5d4e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4",
			Bucket:               "pending-bucket",
			ObjectKey:            "c5/hash.pdf",
			URL:                  "https://example.com/pending.pdf",
			OwnerID:              11111,
			AuthenticationStatus: models.AuthenticationStatusAuthenticating,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		response := presenter.ToDocumentResponse(doc)

		assert.NotNil(t, response)
		assert.Equal(t, "authenticating", response.AuthenticationStatus)
	})
}

func TestToDocumentResponseList(t *testing.T) {
	t.Run("list with multiple documents", func(t *testing.T) {
		now := time.Now()
		docs := []*models.Document{
			{
				ID:                   "doc-1",
				Filename:             "file1.pdf",
				MimeType:             "application/pdf",
				SizeBytes:            1000,
				HashSHA256:           "hash1hash1hash1hash1hash1hash1hash1hash1hash1hash1hash1hash1hash1",
				Bucket:               "bucket1",
				ObjectKey:            "key1",
				URL:                  "https://example.com/file1.pdf",
				OwnerID:              1,
				AuthenticationStatus: models.AuthenticationStatusUnauthenticated,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				ID:                   "doc-2",
				Filename:             "file2.pdf",
				MimeType:             "application/pdf",
				SizeBytes:            2000,
				HashSHA256:           "hash2hash2hash2hash2hash2hash2hash2hash2hash2hash2hash2hash2hash2",
				Bucket:               "bucket2",
				ObjectKey:            "key2",
				URL:                  "https://example.com/file2.pdf",
				OwnerID:              2,
				AuthenticationStatus: models.AuthenticationStatusAuthenticated,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}

		response := presenter.ToDocumentResponseList(docs)

		assert.NotNil(t, response)
		assert.Len(t, response, 2)
		assert.Equal(t, "doc-1", response[0].ID)
		assert.Equal(t, "doc-2", response[1].ID)
	})

	t.Run("empty list", func(t *testing.T) {
		docs := []*models.Document{}
		response := presenter.ToDocumentResponseList(docs)

		assert.NotNil(t, response)
		assert.Len(t, response, 0)
	})

	t.Run("nil list", func(t *testing.T) {
		response := presenter.ToDocumentResponseList(nil)

		assert.NotNil(t, response)
		assert.Len(t, response, 0)
	})

	t.Run("list with one document", func(t *testing.T) {
		docs := []*models.Document{
			{
				ID:                   "doc-single",
				Filename:             "single.pdf",
				MimeType:             "application/pdf",
				SizeBytes:            500,
				HashSHA256:           "singlehashsinglehashsinglehashsinglehashsinglehashsinglehashsi",
				Bucket:               "single-bucket",
				ObjectKey:            "single-key",
				URL:                  "https://example.com/single.pdf",
				OwnerID:              999,
				AuthenticationStatus: models.AuthenticationStatusAuthenticating,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			},
		}

		response := presenter.ToDocumentResponseList(docs)

		assert.NotNil(t, response)
		assert.Len(t, response, 1)
		assert.Equal(t, "doc-single", response[0].ID)
	})

	t.Run("list with nil document is skipped", func(t *testing.T) {
		docs := []*models.Document{
			{
				ID:                   "doc-valid",
				Filename:             "valid.pdf",
				MimeType:             "application/pdf",
				SizeBytes:            100,
				HashSHA256:           "validhashvalidhashvalidhashvalidhashvalidhashvalidhashvalidha",
				Bucket:               "valid-bucket",
				ObjectKey:            "valid-key",
				URL:                  "https://example.com/valid.pdf",
				OwnerID:              123,
				AuthenticationStatus: models.AuthenticationStatusUnauthenticated,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			},
			nil, // This should be skipped
		}

		response := presenter.ToDocumentResponseList(docs)

		assert.NotNil(t, response)
		assert.Len(t, response, 1)
		assert.Equal(t, "doc-valid", response[0].ID)
	})
}
