package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/request"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/presenter"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentUploadHandler struct {
	service      usecases.DocumentService
	errorHandler *errors.ErrorHandler
}

func NewDocumentUploadHandler(service usecases.DocumentService, errorHandler *errors.ErrorHandler) *DocumentUploadHandler {
	return &DocumentUploadHandler{
		service:      service,
		errorHandler: errorHandler,
	}
}

func (handler *DocumentUploadHandler) Upload(ctx *gin.Context) {
	var uploadRequest request.UploadRequest

	if err := ctx.ShouldBind(&uploadRequest); err != nil {
		handler.errorHandler.HandleError(ctx, errors.NewValidationError("invalid request format or validation failed"))
		return
	}

	document, err := handler.service.Upload(ctx.Request.Context(), uploadRequest.File, uploadRequest.Email)
	if err != nil {
		handler.errorHandler.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    presenter.ToDocumentResponse(document),
	})
}