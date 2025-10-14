package domain

// DocumentAuthenticationRequestedEvent represents the event published when a document authentication is requested
type DocumentAuthenticationRequestedEvent struct {
	IDCitizen     string `json:"idCitizen"`     // Owner's email (citizen identifier)
	URLDocument   string `json:"UrlDocument"`   // Pre-signed URL to access the document
	DocumentTitle string `json:"documentTitle"` // Document filename
}
