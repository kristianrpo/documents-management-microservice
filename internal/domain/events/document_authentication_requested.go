package events

// DocumentAuthenticationRequestedEvent represents the event published when a document authentication is requested
type DocumentAuthenticationRequestedEvent struct {
	IDCitizen     int64  `json:"idCitizen"`     // Owner's ID (citizen identifier)
	URLDocument   string `json:"urlDocument"`   // Pre-signed URL to access the document
	DocumentTitle string `json:"documentTitle"` // Document filename
}
