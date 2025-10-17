package models

import (
	"strings"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// Document represents a file uploaded to the system with its metadata
type Document struct {
	ID                   string               `dynamodbav:"DocumentID" json:"id"`                              // Unique document identifier (UUID)
	Filename             string               `dynamodbav:"Filename" json:"filename"`                          // Original filename
	MimeType             string               `dynamodbav:"MimeType" json:"mime_type"`                         // MIME type (e.g., application/pdf)
	SizeBytes            int64                `dynamodbav:"SizeBytes" json:"size_bytes"`                       // File size in bytes
	HashSHA256           string               `dynamodbav:"HashSHA256" json:"hash_sha256"`                     // SHA256 hash for deduplication
	Bucket               string               `dynamodbav:"Bucket" json:"bucket"`                              // S3 bucket name
	ObjectKey            string               `dynamodbav:"ObjectKey" json:"object_key"`                       // S3 object key (path)
	URL                  string               `dynamodbav:"URL" json:"url"`                                    // Public URL (if available)
	OwnerID              int64                `dynamodbav:"OwnerID" json:"owner_id"`                           // Citizen ID who owns the document
	AuthenticationStatus AuthenticationStatus `dynamodbav:"AuthenticationStatus" json:"authentication_status"` // Current authentication state
	CreatedAt            time.Time            `dynamodbav:"CreatedAt" json:"created_at"`                       // Document creation timestamp
	UpdatedAt            time.Time            `dynamodbav:"UpdatedAt" json:"updated_at"`                       // Last update timestamp
}

// Validate checks if the document has all required fields with valid values
func (d *Document) Validate() error {
	if strings.TrimSpace(d.Filename) == "" {
		return errors.NewValidationError("filename cannot be empty")
	}

	if d.SizeBytes <= 0 {
		return errors.NewValidationError("file size must be greater than zero")
	}

	if len(d.HashSHA256) != 64 {
		return errors.NewValidationError("invalid SHA256 hash format (expected 64 characters)")
	}

	if d.OwnerID <= 0 {
		return errors.NewValidationError("owner ID must be greater than zero")
	}

	if strings.TrimSpace(d.ObjectKey) == "" {
		return errors.NewValidationError("object key cannot be empty")
	}

	if strings.TrimSpace(d.Bucket) == "" {
		return errors.NewValidationError("bucket name cannot be empty")
	}

	if d.AuthenticationStatus != "" && !d.AuthenticationStatus.IsValid() {
		return errors.NewValidationError("invalid authentication status")
	}

	return nil
}
