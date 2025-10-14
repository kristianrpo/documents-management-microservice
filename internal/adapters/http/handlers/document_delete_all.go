package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentDeleteAllHandler struct {
	service      usecases.DocumentDeleteAllService
	errorHandler *errors.ErrorHandler
}

func NewDocumentDeleteAllHandler(service usecases.DocumentDeleteAllService, errorHandler *errors.ErrorHandler) *DocumentDeleteAllHandler {
	return &DocumentDeleteAllHandler{
		service:      service,
		errorHandler: errorHandler,
	}
}

// DeleteAll godoc
// @Summary Delete all documents for a user
// @Description Deletes all documents belonging to a specific user (identified by email) and their associated files from S3 storage.
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
// @Description - `VALIDATION_ERROR`: Invalid email format
// @Description - `PERSISTENCE_ERROR`: Failed to delete documents from database
// @Tags documents
// @Accept json
// @Produce json
// @Param email path string true "Owner's email address" format(email) example(user@example.com)
// @Success 200 {object} endpoints.DeleteAllResponse "All documents deleted successfully"
// @Failure 400 {object} endpoints.DeleteAllErrorResponse "Validation error - invalid email format"
// @Failure 500 {object} endpoints.DeleteAllErrorResponse "Internal server error - database or storage error"
// @Router /api/v1/documents/user/{email} [delete]
func (handler *DocumentDeleteAllHandler) DeleteAll(ctx *gin.Context) {
	email := ctx.Param("email")

	if email == "" {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("email is required"))
		return
	}

	deletedCount, err := handler.service.DeleteAll(ctx.Request.Context(), email)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	response := endpoints.DeleteAllResponse{
		Success: true,
		Message: "all documents deleted successfully",
		Data: endpoints.DeleteAllData{
			DeletedCount: deletedCount,
		},
	}

	ctx.JSON(http.StatusOK, response)
}
