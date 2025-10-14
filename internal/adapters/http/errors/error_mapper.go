package errors

import (
	"net/http"

	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type ErrorMapper struct{}

func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

func (m *ErrorMapper) MapDomainErrorToHTTPStatus(err *domain.DomainError) int {
	switch err.Code {
	case domain.ErrCodeValidation:
		return http.StatusBadRequest
	case domain.ErrCodeFileRead:
		return http.StatusBadRequest
	case domain.ErrCodeHashCalculate:
		return http.StatusInternalServerError
	case domain.ErrCodeStorageUpload:
		return http.StatusInternalServerError
	case domain.ErrCodePersistence:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
