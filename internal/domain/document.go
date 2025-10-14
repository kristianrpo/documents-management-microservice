package domain

import (
	"strings"
	"time"
)

type Document struct {
	ID         string    `dynamodbav:"id" json:"id"`
	Filename   string    `dynamodbav:"filename" json:"filename"`
	MimeType   string    `dynamodbav:"mime_type" json:"mime_type"`
	SizeBytes  int64     `dynamodbav:"size_bytes" json:"size_bytes"`
	HashSHA256 string    `dynamodbav:"hash_sha256" json:"hash_sha256"`
	Bucket     string    `dynamodbav:"bucket" json:"bucket"`
	ObjectKey  string    `dynamodbav:"object_key" json:"object_key"`
	URL        string    `dynamodbav:"url" json:"url"`
	OwnerEmail string    `dynamodbav:"owner_email" json:"owner_email"`
	CreatedAt  time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt  time.Time `dynamodbav:"updated_at" json:"updated_at"`
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
	
	if strings.TrimSpace(d.OwnerEmail) == "" {
		return NewValidationError("owner email cannot be empty")
	}
	
	if !isValidEmail(d.OwnerEmail) {
		return NewValidationError("invalid email format")
	}
	
	if strings.TrimSpace(d.ObjectKey) == "" {
		return NewValidationError("object key cannot be empty")
	}
	
	return nil
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	return len(parts[0]) > 0 && len(parts[1]) > 0 && strings.Contains(parts[1], ".")
}
