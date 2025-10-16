package errors

// DomainError represents a domain-level error with a code and optional wrapped error
type DomainError struct {
	Code    string // Error code for categorization
	Message string // Human-readable error message
	Err     error  // Wrapped underlying error (if any)
}

// Error implements the error interface for DomainError
func (e *DomainError) Error() string { return e.Message }

// Unwrap returns the wrapped underlying error if any
func (e *DomainError) Unwrap() error { return e.Err }

// Error codes for different domain error types
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeFileRead      = "FILE_READ_ERROR"
	ErrCodeHashCalculate = "HASH_CALCULATE_ERROR"
	ErrCodeStorageUpload = "STORAGE_UPLOAD_ERROR"
	ErrCodePersistence   = "PERSISTENCE_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
)

// NewValidationError creates a validation error (e.g., invalid input data)
func NewValidationError(message string) *DomainError {
	return &DomainError{Code: ErrCodeValidation, Message: message}
}

// NewFileReadError creates an error when file reading fails
func NewFileReadError(err error) *DomainError {
	return &DomainError{Code: ErrCodeFileRead, Message: "failed to read file", Err: err}
}

// NewHashCalculateError creates an error when hash calculation fails
func NewHashCalculateError(err error) *DomainError {
	return &DomainError{Code: ErrCodeHashCalculate, Message: "failed to calculate file hash", Err: err}
}

// NewStorageUploadError creates an error when storage upload fails
func NewStorageUploadError(err error) *DomainError {
	return &DomainError{Code: ErrCodeStorageUpload, Message: "failed to upload to storage", Err: err}
}

// NewPersistenceError creates an error when database operations fail
func NewPersistenceError(err error) *DomainError {
	return &DomainError{Code: ErrCodePersistence, Message: "failed to persist document", Err: err}
}

// NewNotFoundError creates an error when a resource is not found
func NewNotFoundError(message string) *DomainError {
	return &DomainError{Code: ErrCodeNotFound, Message: message}
}
