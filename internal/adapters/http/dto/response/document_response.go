package response

type DocumentResponse struct {
	ID         uint   `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	HashSHA256 string `json:"hash_sha256"`
	URL        string `json:"url"`
	OwnerEmail string `json:"owner_email"`
}
