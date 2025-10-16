package errors

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
)

// ErrorHandler handles error responses for HTTP requests
type ErrorHandler struct {
	mapper *ErrorMapper
}

// NewErrorHandler creates a new error handler with the provided error mapper
func NewErrorHandler(mapper *ErrorMapper) *ErrorHandler {
	return &ErrorHandler{mapper: mapper}
}

// HandleError processes errors and sends appropriate HTTP responses based on error type
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
		if domainErr.Err != nil {
			log.Printf("domain error %s: %v", domainErr.Code, domainErr.Err)
		}
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
