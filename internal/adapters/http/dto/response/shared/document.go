package shared

type DocumentResponse struct {
	ID         string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Filename   string `json:"filename" example:"my-document.pdf"`
	MimeType   string `json:"mime_type" example:"application/pdf"`
	SizeBytes  int64  `json:"size_bytes" example:"102400"`
	HashSHA256 string `json:"hash_sha256" example:"abc123def456789..."`
	URL        string `json:"url" example:"https://my-bucket.s3.amazonaws.com/ab/abc123def456.pdf"`
	OwnerEmail string `json:"owner_email" example:"user@example.com"`
}
