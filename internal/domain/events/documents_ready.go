package events

// DocumentsReadyEvent is published when batch download/upload operation completes
type DocumentsReadyEvent struct {
	IDCitizen int64  `json:"idCitizen"`
	Status    string `json:"status"` // "success" or "failure"
	Message   string `json:"message,omitempty"`
}
