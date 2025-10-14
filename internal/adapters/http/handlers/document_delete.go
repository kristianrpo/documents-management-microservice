package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentDeleteHandler struct {
	service      usecases.DocumentDeleteService
	errorHandler *errors.ErrorHandler
}

func NewDocumentDeleteHandler(service usecases.DocumentDeleteService, errorHandler *errors.ErrorHandler) *DocumentDeleteHandler {
	return &DocumentDeleteHandler{
		service:      service,
		errorHandler: errorHandler,
	}
}

// Delete godoc
// @Summary Delete a document by ID
// @Description Deletes a document and its associated file from S3 storage.
// @Description
// @Description ## Features
// @Description - Deletes document metadata from DynamoDB
// @Description - Removes the physical file from S3 storage
// @Description - Returns 404 if document doesn't exist
// @Description
// @Description ## Use Cases
// @Description - Remove unwanted documents
// @Description - Clean up storage space
// @Description - Comply with data deletion requests
// @Description
// @Description ## Error Codes
// @Description - `NOT_FOUND`: Document with the specified ID does not exist
// @Description - `PERSISTENCE_ERROR`: Failed to delete document from database
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} endpoints.DeleteResponse "Document deleted successfully"
// @Failure 404 {object} endpoints.DeleteErrorResponse "Document not found"
// @Failure 500 {object} endpoints.DeleteErrorResponse "Internal server error - database or storage error"
// @Router /api/v1/documents/{id} [delete]
func (handler *DocumentDeleteHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("document id is required"))
		return
	}

	err := handler.service.Delete(ctx.Request.Context(), id)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	response := endpoints.DeleteResponse{
		Success: true,
		Message: "document deleted successfully",
	}

	ctx.JSON(http.StatusOK, response)
}
