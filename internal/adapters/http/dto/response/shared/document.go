package shared

type DocumentResponse struct {
	ID                   string `json:"id"`
	Filename             string `json:"filename"`
	MimeType             string `json:"mime_type"`
	SizeBytes            int64  `json:"size_bytes"`
	HashSHA256           string `json:"hash_sha256"`
	URL                  string `json:"url"`
	OwnerID              int64  `json:"owner_id"`
	AuthenticationStatus string `json:"authentication_status"`
}
