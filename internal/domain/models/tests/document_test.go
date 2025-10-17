package models_test

import (
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestDocument_Validate(t *testing.T) {
	validHash := "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	tests := []struct {
		name        string
		document    *models.Document
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid document",
			document: &models.Document{
				ID:                   "doc-123",
				Filename:             "test.pdf",
				MimeType:             "application/pdf",
				SizeBytes:            1024,
				HashSHA256:           validHash,
				Bucket:               "test-bucket",
				ObjectKey:            "test/path/file.pdf",
				URL:                  "https://example.com/file.pdf",
				OwnerID:              12345,
				AuthenticationStatus: models.AuthenticationStatusUnauthenticated,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty filename",
			document: &models.Document{
				Filename:   "",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "filename cannot be empty",
		},
		{
			name: "whitespace filename",
			document: &models.Document{
				Filename:   "   ",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "filename cannot be empty",
		},
		{
			name: "zero file size",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  0,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "file size must be greater than zero",
		},
		{
			name: "negative file size",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  -100,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "file size must be greater than zero",
		},
		{
			name: "invalid hash too short",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: "tooshort",
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "invalid SHA256 hash format (expected 64 characters)",
		},
		{
			name: "invalid hash too long",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash + "extra",
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "invalid SHA256 hash format (expected 64 characters)",
		},
		{
			name: "zero owner ID",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    0,
			},
			expectError: true,
			errorMsg:    "owner ID must be greater than zero",
		},
		{
			name: "negative owner ID",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "test/path",
				OwnerID:    -1,
			},
			expectError: true,
			errorMsg:    "owner ID must be greater than zero",
		},
		{
			name: "empty object key",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "object key cannot be empty",
		},
		{
			name: "whitespace object key",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "test-bucket",
				ObjectKey:  "   ",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "object key cannot be empty",
		},
		{
			name: "empty bucket",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "bucket name cannot be empty",
		},
		{
			name: "whitespace bucket",
			document: &models.Document{
				Filename:   "test.pdf",
				SizeBytes:  1024,
				HashSHA256: validHash,
				Bucket:     "   ",
				ObjectKey:  "test/path",
				OwnerID:    12345,
			},
			expectError: true,
			errorMsg:    "bucket name cannot be empty",
		},
		{
			name: "invalid authentication status",
			document: &models.Document{
				Filename:             "test.pdf",
				SizeBytes:            1024,
				HashSHA256:           validHash,
				Bucket:               "test-bucket",
				ObjectKey:            "test/path",
				OwnerID:              12345,
				AuthenticationStatus: models.AuthenticationStatus("invalid"),
			},
			expectError: true,
			errorMsg:    "invalid authentication status",
		},
		{
			name: "empty authentication status is valid",
			document: &models.Document{
				Filename:             "test.pdf",
				SizeBytes:            1024,
				HashSHA256:           validHash,
				Bucket:               "test-bucket",
				ObjectKey:            "test/path",
				OwnerID:              12345,
				AuthenticationStatus: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.document.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
