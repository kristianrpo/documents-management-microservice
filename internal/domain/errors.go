package domain

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string { return e.Message }
func (e *DomainError) Unwrap() error { return e.Err }

const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeFileRead      = "FILE_READ_ERROR"
	ErrCodeHashCalculate = "HASH_CALCULATE_ERROR"
	ErrCodeStorageUpload = "STORAGE_UPLOAD_ERROR"
	ErrCodePersistence   = "PERSISTENCE_ERROR"
)

func NewValidationError(message string) *DomainError {
	return &DomainError{Code: ErrCodeValidation, Message: message}
}

func NewFileReadError(err error) *DomainError {
	return &DomainError{Code: ErrCodeFileRead, Message: "failed to read file", Err: err}
}

func NewHashCalculateError(err error) *DomainError {
	return &DomainError{Code: ErrCodeHashCalculate, Message: "failed to calculate file hash", Err: err}
}

func NewStorageUploadError(err error) *DomainError {
	return &DomainError{Code: ErrCodeStorageUpload, Message: "failed to upload to storage", Err: err}
}

func NewPersistenceError(err error) *DomainError {
	return &DomainError{Code: ErrCodePersistence, Message: "failed to persist document", Err: err}
}
