package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentRequestAuthenticationHandler struct {
	authService  usecases.DocumentRequestAuthenticationService
	errorHandler *errors.ErrorHandler
}

func NewDocumentRequestAuthenticationHandler(
	authService usecases.DocumentRequestAuthenticationService,
	errorHandler *errors.ErrorHandler,
) *DocumentRequestAuthenticationHandler {
	return &DocumentRequestAuthenticationHandler{
		authService:  authService,
		errorHandler: errorHandler,
	}
}

// RequestAuthentication godoc
// @Summary Request document authentication
// @Description Requests authentication of a document by publishing an event for external authentication service. The document owner's citizen ID and filename are automatically included in the event.
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Success 202 {object} endpoints.RequestAuthenticationResponse "Authentication request accepted"
// @Failure 400 {object} endpoints.RequestAuthenticationErrorResponse "Invalid request"
// @Failure 404 {object} endpoints.RequestAuthenticationErrorResponse "Document not found"
// @Failure 500 {object} endpoints.RequestAuthenticationErrorResponse "Internal server error"
// @Router /api/v1/documents/{id}/request-authentication [post]
func (h *DocumentRequestAuthenticationHandler) RequestAuthentication(c *gin.Context) {
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, endpoints.RequestAuthenticationErrorResponse{
			Success: false,
			Error: shared.ErrorDetail{
				Code:    "INVALID_DOCUMENT_ID",
				Message: "Document ID parameter is required",
			},
		})
		return
	}

	log.Printf("Requesting authentication for document %s", documentID)

	err := h.authService.RequestAuthentication(c.Request.Context(), documentID)
	if err != nil {
		h.errorHandler.HandleError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, endpoints.RequestAuthenticationResponse{
		Success: true,
		Message: "Authentication request submitted successfully",
	})
}
