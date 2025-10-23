package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// DocumentListHandler handles HTTP requests for listing documents
type DocumentListHandler struct {
	service      usecases.DocumentListService
	errorHandler *errors.ErrorHandler
	metrics      *metrics.PrometheusMetrics
}

// NewDocumentListHandler creates a new handler for document list operations
func NewDocumentListHandler(service usecases.DocumentListService, errorHandler *errors.ErrorHandler, metrics *metrics.PrometheusMetrics) *DocumentListHandler {
	return &DocumentListHandler{
		service:      service,
		errorHandler: errorHandler,
		metrics:      metrics,
	}
}

// List godoc
// @Summary List documents
// @Description Retrieves a paginated list of documents for a specific owner (identified by citizen ID).
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
// @Description - `VALIDATION_ERROR`: Invalid id_citizen or pagination parameters
// @Description - `PERSISTENCE_ERROR`: Failed to retrieve documents from database
// @Tags documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (starts at 1)" minimum(1) default(1) example(1)
// @Param limit query int false "Number of items per page (max 100)" minimum(1) maximum(100) default(10) example(10)
// @Success 200 {object} endpoints.ListResponse "List of documents retrieved successfully"
// @Failure 400 {object} endpoints.ListErrorResponse "Validation error - invalid id_citizen or pagination parameters"
// @Failure 500 {object} endpoints.ListErrorResponse "Internal server error - database error"
// @Router /api/v1/documents [get]
func (handler *DocumentListHandler) List(ctx *gin.Context) {
	idCitizen, err := middleware.GetUserIDCitizen(ctx)
	if err != nil {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("user not authenticated"))
		return
	}

	page := 1
	limit := 10
	if p := ctx.Query("page"); p != "" {
		_, _ = fmt.Sscanf(p, "%d", &page)
	}
	if l := ctx.Query("limit"); l != "" {
		_, _ = fmt.Sscanf(l, "%d", &limit)
	}

	documents, pagination, totalPages, totalCount, err := handler.service.List(
		ctx.Request.Context(),
		idCitizen,
		page,
		limit,
	)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	handler.metrics.ListRequestsTotal.Inc()

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
