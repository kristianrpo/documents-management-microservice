package events

// UserTransferredEvent represents the event when a user is transferred to another operator
// All documents owned by this citizen should be deleted as part of the transfer process
type UserTransferredEvent struct {
	IDCitizen int64 `json:"idCitizen"` // Citizen ID being transferred
}
