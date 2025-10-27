package events

// DocumentAuthenticationRequestedEvent represents the event published when a document authentication is requested
type DocumentAuthenticationRequestedEvent struct {
	MessageID     string `json:"messageId"`     // Unique message ID for deduplication
	IDCitizen     int64  `json:"idCitizen"`     // Owner's ID (citizen identifier)
	URLDocument   string `json:"urlDocument"`   // Pre-signed URL to access the document
	DocumentTitle string `json:"documentTitle"` // Document filename
	DocumentID    string `json:"documentId"`    // Document ID to track the authentication result
}
