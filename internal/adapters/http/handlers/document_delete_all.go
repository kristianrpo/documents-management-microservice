package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// DocumentDeleteAllHandler handles HTTP requests for bulk document deletion for a specific user
type DocumentDeleteAllHandler struct {
	service      usecases.DocumentDeleteAllService
	errorHandler *errors.ErrorHandler
	metrics      *metrics.PrometheusMetrics
}

// NewDocumentDeleteAllHandler creates a new handler for bulk document deletion operations
func NewDocumentDeleteAllHandler(service usecases.DocumentDeleteAllService, errorHandler *errors.ErrorHandler, metrics *metrics.PrometheusMetrics) *DocumentDeleteAllHandler {
	return &DocumentDeleteAllHandler{
		service:      service,
		errorHandler: errorHandler,
		metrics:      metrics,
	}
}

// DeleteAll godoc
// @Summary Delete all documents for a user
// @Description Deletes all documents belonging to a specific user (identified by citizen ID) and their associated files from S3 storage.
// @Description
// @Description ## Features
// @Description - Deletes all document metadata from DynamoDB for the specified user
// @Description - Removes all physical files from S3 storage
// @Description - Returns the count of deleted documents
// @Description - Useful for account closure or data migration scenarios
// @Description
// @Description ## Use Cases
// @Description - Account closure/deletion
// @Description - Data migration to another system
// @Description - Bulk cleanup operations
// @Description - GDPR/privacy compliance (right to be forgotten)
// @Description
// @Description ## Error Codes
// @Description - `VALIDATION_ERROR`: Invalid citizen ID
// @Description - `PERSISTENCE_ERROR`: Failed to delete documents from database
// @Tags documents
// @Accept json
// @Produce json
// @Param id_citizen path int true "Citizen ID" example(123456789)
// @Success 200 {object} endpoints.DeleteAllResponse "All documents deleted successfully"
// @Failure 400 {object} endpoints.DeleteAllErrorResponse "Validation error - invalid citizen ID"
// @Failure 500 {object} endpoints.DeleteAllErrorResponse "Internal server error - database or storage error"
// @Router /api/v1/documents/user/{id_citizen} [delete]
func (handler *DocumentDeleteAllHandler) DeleteAll(ctx *gin.Context) {
	idCitizenStr := ctx.Param("id_citizen")

	idCitizen, err := strconv.ParseInt(idCitizenStr, 10, 64)
	if err != nil || idCitizen <= 0 {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("id_citizen must be a valid positive integer"))
		return
	}

	deletedCount, err := handler.service.DeleteAll(ctx.Request.Context(), idCitizen)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	// Increment bulk delete metric with the count
	handler.metrics.DeleteBulkRequestsTotal.Inc()

	response := endpoints.DeleteAllResponse{
		Success: true,
		Message: "all documents deleted successfully",
		Data: endpoints.DeleteAllData{
			DeletedCount: deletedCount,
		},
	}

	ctx.JSON(http.StatusOK, response)
}
