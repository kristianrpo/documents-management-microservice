package events

// DocumentAuthenticationCompletedEvent represents the event received when a document has been authenticated
// This event is published by the operator-connectivity microservice after calling the authentication service
type DocumentAuthenticationCompletedEvent struct {
	MessageID      string `json:"messageId"`      // Unique message ID for deduplication (from original request)
	DocumentID     string `json:"documentId"`      // Document ID that was authenticated
	IDCitizen      int64  `json:"idCitizen"`       // Owner's ID (citizen identifier)
	Authenticated  bool   `json:"authenticated"`   // Whether the authentication was successful
	Message        string `json:"message"`         // Authentication result message
	AuthenticatedAt string `json:"authenticatedAt"` // Timestamp when authentication completed (ISO 8601)
}
