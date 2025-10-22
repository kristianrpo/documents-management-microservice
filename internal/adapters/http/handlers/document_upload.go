package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// DocumentUploadHandler handles HTTP requests for document upload operations
type DocumentUploadHandler struct {
	service      usecases.DocumentService
	errorHandler *errors.ErrorHandler
	metrics      *metrics.PrometheusMetrics
}

func NewDocumentUploadHandler(service usecases.DocumentService, errorHandler *errors.ErrorHandler, metricsCollector *metrics.PrometheusMetrics) *DocumentUploadHandler {
	return &DocumentUploadHandler{
		service:      service,
		errorHandler: errorHandler,
		metrics:      metricsCollector,
	}
}

// Upload godoc
// @Summary Upload a document
// @Description Uploads a document to S3 storage and saves its metadata. The owner is determined from JWT token.
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 201 {object} endpoints.UploadResponse "Document uploaded successfully"
// @Failure 400 {object} endpoints.UploadErrorResponse "Validation error"
// @Failure 401 {object} endpoints.UploadErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} endpoints.UploadErrorResponse "Internal server error"
// @Router /api/v1/documents [post]
func (handler *DocumentUploadHandler) Upload(ctx *gin.Context) {
	// Get user ID from JWT token
	idCitizen, err := middleware.GetUserIDCitizen(ctx)
	if err != nil {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("user not authenticated"))
		return
	}

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("file is required"))
		return
	}

	document, err := handler.service.Upload(ctx.Request.Context(), file, idCitizen)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	handler.metrics.UploadRequestsTotal.Inc()

	ctx.JSON(http.StatusCreated, endpoints.UploadResponse{
		Success: true,
		Data:    *presenter.ToDocumentResponse(document),
	})
}
