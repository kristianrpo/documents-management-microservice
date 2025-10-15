package errors

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

type ErrorHandler struct {
	mapper *ErrorMapper
}

func NewErrorHandler(mapper *ErrorMapper) *ErrorHandler {
	return &ErrorHandler{mapper: mapper}
}

func (h *ErrorHandler) HandleError(ctx *gin.Context, err error) {
	var domainErr *domainerrors.DomainError
	var validationErr *ValidationError
	debug := os.Getenv("DEBUG") == "true"

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
		// Log underlying error details to server logs for debugging/observability
		if domainErr.Err != nil {
			log.Printf("domain error %s: %v", domainErr.Code, domainErr.Err)
		}
		// Build error payload and include details in debug mode
		payload := gin.H{
			"success": false,
			"error": gin.H{
				"code":    domainErr.Code,
				"message": domainErr.Message,
			},
		}
		if debug && domainErr.Err != nil {
			payload["error"].(gin.H)["details"] = domainErr.Err.Error()
		}
		ctx.JSON(statusCode, payload)
	default:
		// Log unexpected error
		if err != nil {
			log.Printf("internal error: %v", err)
		}
		payload := gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "an unexpected error occurred",
			},
		}
		if debug && err != nil {
			payload["error"].(gin.H)["details"] = err.Error()
		}
		ctx.JSON(http.StatusInternalServerError, payload)
	}
}
