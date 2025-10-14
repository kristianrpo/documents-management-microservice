package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/request"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
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

// Upload godoc
// @Summary Upload a document
// @Description Uploads a document to S3 storage and saves its metadata in DynamoDB.
// @Description
// @Description ## Features
// @Description - Automatic file deduplication based on SHA256 hash
// @Description - If a file with the same hash exists for the same user, returns the existing document
// @Description - Supports any file type
// @Description - Automatically detects MIME type from file extension
// @Description - Generates unique object keys using hash prefix for optimal S3 performance
// @Description
// @Description ## Process
// @Description 1. Calculates SHA256 hash of the uploaded file
// @Description 2. Checks if document already exists (hash + email)
// @Description 3. If exists, returns existing document (no duplicate upload)
// @Description 4. If new, uploads to S3 and saves metadata to DynamoDB
// @Description
// @Description ## Error Codes
// @Description - `VALIDATION_ERROR`: Invalid request format or missing required fields
// @Description - `FILE_READ_ERROR`: Failed to read the uploaded file
// @Description - `HASH_CALCULATE_ERROR`: Failed to calculate file hash
// @Description - `STORAGE_UPLOAD_ERROR`: Failed to upload file to S3
// @Description - `PERSISTENCE_ERROR`: Failed to save metadata to DynamoDB
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload (supports any file type: PDF, images, documents, etc.)"
// @Param email formData string true "Owner's email address" format(email) example(user@example.com)
// @Success 201 {object} endpoints.DocumentUploadSuccessResponse "Document uploaded successfully"
// @Failure 400 {object} endpoints.DocumentUploadErrorResponse "Validation error - invalid email format or missing required fields"
// @Failure 500 {object} endpoints.DocumentUploadErrorResponse "Internal server error - file processing, storage upload, or database error"
// @Router /api/v1/documents [post]
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

	ctx.JSON(http.StatusCreated, endpoints.DocumentUploadSuccessResponse{
		Success: true,
		Data:    *presenter.ToDocumentData(document),
	})
}
