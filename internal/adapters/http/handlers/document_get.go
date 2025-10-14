package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentGetHandler struct {
	service      usecases.DocumentGetService
	errorHandler *errors.ErrorHandler
}

func NewDocumentGetHandler(service usecases.DocumentGetService, errorHandler *errors.ErrorHandler) *DocumentGetHandler {
	return &DocumentGetHandler{
		service:      service,
		errorHandler: errorHandler,
	}
}

// GetByID godoc
// @Summary Get document by ID
// @Description Retrieves detailed information about a specific document by its ID.
// @Description
// @Description ## Features
// @Description - Returns complete document metadata including URL for viewing/downloading
// @Description - URL is pre-signed and ready to use in frontend viewers
// @Description - Includes file information (size, type, hash, etc.)
// @Description
// @Description ## Use Cases
// @Description - Display document details in UI
// @Description - Preview documents in viewers (PDF, images, etc.)
// @Description - Download documents
// @Description - Verify document integrity using hash
// @Description
// @Description ## Error Codes
// @Description - `NOT_FOUND`: Document with the specified ID does not exist
// @Description - `PERSISTENCE_ERROR`: Failed to retrieve document from database
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} endpoints.GetResponse "Document retrieved successfully"
// @Failure 404 {object} endpoints.GetErrorResponse "Document not found"
// @Failure 500 {object} endpoints.GetErrorResponse "Internal server error - database error"
// @Router /api/v1/documents/{id} [get]
func (handler *DocumentGetHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("document id is required"))
		return
	}

	document, err := handler.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	response := endpoints.GetResponse{
		Success: true,
		Data:    *presenter.ToDocumentResponse(document),
	}

	ctx.JSON(http.StatusOK, response)
}
