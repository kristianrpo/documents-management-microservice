package errors

// ValidationError represents an error that occurs during request validation
type ValidationError struct {
	Message string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error with the provided message
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}
