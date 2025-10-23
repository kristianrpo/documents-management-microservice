package events

// DocumentDownloadRequestedEvent represents an event sent to request the service
// to download a set of pre-signed URLs and store them as documents for a user.
type DocumentDownloadRequestedEvent struct {
	IDCitizen int64    `json:"idCitizen"`
	URLs      []string `json:"urls"`
}
