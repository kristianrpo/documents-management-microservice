package errors

import (
	"net/http"

	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// ErrorMapper maps domain errors to HTTP status codes
type ErrorMapper struct{}

// NewErrorMapper creates a new error mapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapDomainErrorToHTTPStatus maps a domain error to its corresponding HTTP status code
func (m *ErrorMapper) MapDomainErrorToHTTPStatus(err *domainerrors.DomainError) int {
	switch err.Code {
	case domainerrors.ErrCodeValidation:
		return http.StatusBadRequest
	case domainerrors.ErrCodeFileRead:
		return http.StatusBadRequest
	case domainerrors.ErrCodeHashCalculate:
		return http.StatusInternalServerError
	case domainerrors.ErrCodeStorageUpload:
		return http.StatusInternalServerError
	case domainerrors.ErrCodePersistence:
		return http.StatusInternalServerError
	case domainerrors.ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
