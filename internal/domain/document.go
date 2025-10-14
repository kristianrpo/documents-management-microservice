package domain

import (
	"strings"
	"time"
)

type Document struct {
	ID         string    `dynamodbav:"DocumentID" json:"id"`
	Filename   string    `dynamodbav:"Filename" json:"filename"`
	MimeType   string    `dynamodbav:"MimeType" json:"mime_type"`
	SizeBytes  int64     `dynamodbav:"SizeBytes" json:"size_bytes"`
	HashSHA256 string    `dynamodbav:"HashSHA256" json:"hash_sha256"`
	Bucket     string    `dynamodbav:"Bucket" json:"bucket"`
	ObjectKey  string    `dynamodbav:"ObjectKey" json:"object_key"`
	URL        string    `dynamodbav:"URL" json:"url"`
	OwnerID    int64     `dynamodbav:"OwnerID" json:"owner_id"`
	CreatedAt  time.Time `dynamodbav:"CreatedAt" json:"created_at"`
	UpdatedAt  time.Time `dynamodbav:"UpdatedAt" json:"updated_at"`
}

func (d *Document) Validate() error {
	if strings.TrimSpace(d.Filename) == "" {
		return NewValidationError("filename cannot be empty")
	}

	if d.SizeBytes <= 0 {
		return NewValidationError("file size must be greater than zero")
	}

	if len(d.HashSHA256) != 64 {
		return NewValidationError("invalid SHA256 hash format")
	}

	if d.OwnerID <= 0 {
		return NewValidationError("owner ID must be greater than zero")
	}

	if strings.TrimSpace(d.ObjectKey) == "" {
		return NewValidationError("object key cannot be empty")
	}

	return nil
}
