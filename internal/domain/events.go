package domain

// DocumentAuthenticationRequestedEvent represents the event published when a document authentication is requested
type DocumentAuthenticationRequestedEvent struct {
	IDCitizen     int64  `json:"idCitizen"`     // Owner's ID (citizen identifier)
	URLDocument   string `json:"UrlDocument"`   // Pre-signed URL to access the document
	DocumentTitle string `json:"documentTitle"` // Document filename
}

// UserTransferredEvent represents the event when a user is transferred to another operator
type UserTransferredEvent struct {
	IDCitizen int64 `json:"idCitizen"`
}

