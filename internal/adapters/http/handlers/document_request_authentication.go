package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// DocumentRequestAuthenticationHandler handles HTTP requests for requesting document authentication
type DocumentRequestAuthenticationHandler struct {
	authService  usecases.DocumentRequestAuthenticationService
	getService   usecases.DocumentGetService
	errorHandler *errors.ErrorHandler
	metrics      *metrics.PrometheusMetrics
}

// NewDocumentRequestAuthenticationHandler creates a new handler for document authentication request operations
func NewDocumentRequestAuthenticationHandler(
	authService usecases.DocumentRequestAuthenticationService,
	getService usecases.DocumentGetService,
	errorHandler *errors.ErrorHandler,
	metricsCollector *metrics.PrometheusMetrics,
) *DocumentRequestAuthenticationHandler {
	return &DocumentRequestAuthenticationHandler{
		authService:  authService,
		getService:   getService,
		errorHandler: errorHandler,
		metrics:      metricsCollector,
	}
}

// RequestAuthentication godoc
// @Summary Request document authentication
// @Description Requests authentication of a document by publishing an event for external authentication service. The document owner's citizen ID and filename are automatically included in the event.
// @Tags documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Success 202 {object} endpoints.RequestAuthenticationResponse "Authentication request accepted"
// @Failure 400 {object} endpoints.RequestAuthenticationErrorResponse "Invalid request"
// @Failure 404 {object} endpoints.RequestAuthenticationErrorResponse "Document not found"
// @Failure 500 {object} endpoints.RequestAuthenticationErrorResponse "Internal server error"
// @Router /api/v1/documents/{id}/request-authentication [post]
func (h *DocumentRequestAuthenticationHandler) RequestAuthentication(c *gin.Context) {
	// If the authentication service is not available (e.g., RabbitMQ disabled or failed to initialize),
	// return a clear 503 error instead of panicking on nil pointer dereference.
	if h.authService == nil {
		c.JSON(http.StatusServiceUnavailable, endpoints.RequestAuthenticationErrorResponse{
			Success: false,
			Error: shared.ErrorDetail{
				Code:    "SERVICE_UNAVAILABLE",
				Message: "Authentication service is not available. Please try again later.",
			},
		})
		return
	}
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

	document, err := h.getService.GetByID(c.Request.Context(), documentID)
	if err != nil {
		h.errorHandler.HandleError(c, err)
		return
	}
	idCitizen, err := middleware.GetUserIDCitizen(c)
	if document.OwnerID != idCitizen {
		h.errorHandler.HandleError(c, errors.NewValidationError("forbidden: user is not the owner of the document"))
		return
	}

	err = h.authService.RequestAuthentication(c.Request.Context(), documentID)
	if err != nil {
		h.errorHandler.HandleError(c, err)
		return
	}

	if h.metrics != nil && h.metrics.AuthRequestsTotal != nil {
		h.metrics.AuthRequestsTotal.Inc()
	}

	c.JSON(http.StatusAccepted, endpoints.RequestAuthenticationResponse{
		Success: true,
		Message: "Authentication request submitted successfully",
	})
}
