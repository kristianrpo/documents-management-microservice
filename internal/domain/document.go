package domain

import (
	"strings"
	"time"
)

type Document struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	HashSHA256 string    `gorm:"size:64;index" json:"hash_sha256"`
	Bucket     string    `json:"bucket"`
	ObjectKey  string    `json:"object_key"`
	URL        string    `json:"url"`
	OwnerEmail string    `json:"owner_email" gorm:"index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
