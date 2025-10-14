package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/request"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentListHandler struct {
	service      usecases.DocumentListService
	errorHandler *errors.ErrorHandler
}

func NewDocumentListHandler(service usecases.DocumentListService, errorHandler *errors.ErrorHandler) *DocumentListHandler {
	return &DocumentListHandler{
		service:      service,
		errorHandler: errorHandler,
	}
}

// List godoc
// @Summary List documents
// @Description Retrieves a paginated list of documents for a specific owner (identified by email).
// @Description
// @Description ## Features
// @Description - Returns documents sorted by creation date (most recent first)
// @Description - Supports pagination with configurable page size
// @Description - Includes pagination metadata (total items, total pages, current page)
// @Description - Maximum limit per page: 100 documents
// @Description - Default page size: 10 documents
// @Description
// @Description ## Pagination
// @Description - Use `page` parameter to navigate through results (starts at 1)
// @Description - Use `limit` parameter to control page size (1-100)
// @Description - Response includes total count and total pages for UI rendering
// @Description
// @Description ## Error Codes
// @Description - `VALIDATION_ERROR`: Invalid email format or pagination parameters
// @Description - `PERSISTENCE_ERROR`: Failed to retrieve documents from database
// @Tags documents
// @Accept json
// @Produce json
// @Param email query string true "Owner's email address" format(email) example(user@example.com)
// @Param page query int false "Page number (starts at 1)" minimum(1) default(1) example(1)
// @Param limit query int false "Number of items per page (max 100)" minimum(1) maximum(100) default(10) example(10)
// @Success 200 {object} endpoints.ListResponse "List of documents retrieved successfully"
// @Failure 400 {object} endpoints.ListErrorResponse "Validation error - invalid email or pagination parameters"
// @Failure 500 {object} endpoints.ListErrorResponse "Internal server error - database error"
// @Router /api/v1/documents [get]
func (handler *DocumentListHandler) List(ctx *gin.Context) {
	var listRequest request.ListDocumentsRequest

	if err := ctx.ShouldBindQuery(&listRequest); err != nil {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("invalid query parameters"))
		return
	}

	documents, pagination, totalPages, totalCount, err := handler.service.List(
		ctx.Request.Context(),
		listRequest.Email,
		listRequest.Page,
		listRequest.Limit,
	)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	response := endpoints.ListResponse{
		Success: true,
		Data: endpoints.ListData{
			Documents: presenter.ToDocumentResponseList(documents),
			Pagination: shared.Pagination{
				Page:       pagination.Page,
				Limit:      pagination.Limit,
				TotalItems: totalCount,
				TotalPages: totalPages,
			},
		},
	}

	ctx.JSON(http.StatusOK, response)
}
