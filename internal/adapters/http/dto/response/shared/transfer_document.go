package shared

// TransferDocument represents a document with pre-signed URL for transfer
type TransferDocument struct {
	ID           string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Filename     string `json:"filename" example:"passport.pdf"`
	MimeType     string `json:"mime_type" example:"application/pdf"`
	SizeBytes    int64  `json:"size_bytes" example:"1048576"`
	HashSHA256   string `json:"hash_sha256" example:"a1b2c3d4e5f6..."`
	PresignedURL string `json:"presigned_url" example:"https://s3.amazonaws.com/bucket/key?signature=..."`
	ExpiresAt    string `json:"expires_at" example:"2025-10-14T15:30:00Z"`
}
