package errors

import (
	"net/http"

	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

type ErrorMapper struct{}

func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

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
