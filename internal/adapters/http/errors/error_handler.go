package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type ErrorHandler struct {
	mapper *ErrorMapper
}

func NewErrorHandler(mapper *ErrorMapper) *ErrorHandler {
	return &ErrorHandler{mapper: mapper}
}

func (h *ErrorHandler) HandleError(ctx *gin.Context, err error) {
	var domainErr *domain.DomainError
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": validationErr.Message,
			},
		})
	case errors.As(err, &domainErr):
		statusCode := h.mapper.MapDomainErrorToHTTPStatus(domainErr)
		ctx.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    domainErr.Code,
				"message": domainErr.Message,
			},
		})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "an unexpected error occurred",
			},
		})
	}
}
