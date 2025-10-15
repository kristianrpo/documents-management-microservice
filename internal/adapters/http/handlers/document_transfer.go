package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
)

type DocumentTransferHandler struct {
	transferService usecases.DocumentTransferService
	errorHandler    *errors.ErrorHandler
}

func NewDocumentTransferHandler(
	transferService usecases.DocumentTransferService,
	errorHandler *errors.ErrorHandler,
) *DocumentTransferHandler {
	return &DocumentTransferHandler{
		transferService: transferService,
		errorHandler:    errorHandler,
	}
}

// PrepareTransfer godoc
// @Summary Prepare documents for transfer
// @Description Generates pre-signed URLs for all documents owned by a user for secure transfer to another operator
// @Tags documents
// @Accept json
// @Produce json
// @Param id_citizen path int true "Citizen ID"
// @Success 200 {object} endpoints.TransferResponse "Documents prepared successfully"
// @Failure 400 {object} endpoints.TransferErrorResponse "Invalid request"
// @Failure 500 {object} endpoints.TransferErrorResponse "Internal server error"
// @Router /api/v1/documents/transfer/{id_citizen} [get]
func (h *DocumentTransferHandler) PrepareTransfer(c *gin.Context) {
	idCitizenStr := c.Param("id_citizen")
	
	idCitizen, err := strconv.ParseInt(idCitizenStr, 10, 64)
	if err != nil || idCitizen <= 0 {
		c.JSON(http.StatusBadRequest, endpoints.TransferErrorResponse{
			Success: false,
			Error: shared.ErrorDetail{
				Code:    "INVALID_ID_CITIZEN",
				Message: "id_citizen must be a valid positive integer",
			},
		})
		return
	}

	log.Printf("Preparing transfer for user ID: %d", idCitizen)

	results, err := h.transferService.PrepareTransfer(c.Request.Context(), idCitizen)
	if err != nil {
		h.errorHandler.HandleError(c, err)
		return
	}

	// Convert to response DTOs
	transferDocuments := make([]shared.TransferDocument, 0, len(results))
	var expiresAt string

	for _, result := range results {
		if expiresAt == "" {
			expiresAt = result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}

		transferDocuments = append(transferDocuments, shared.TransferDocument{
			ID:           result.Document.ID,
			Filename:     result.Document.Filename,
			MimeType:     result.Document.MimeType,
			SizeBytes:    result.Document.SizeBytes,
			HashSHA256:   result.Document.HashSHA256,
			PresignedURL: result.PresignedURL,
			ExpiresAt:    expiresAt,
		})
	}

	log.Printf("Successfully prepared %d documents for transfer (user ID: %d)", len(transferDocuments), idCitizen)

	c.JSON(http.StatusOK, endpoints.TransferResponse{
		Success: true,
		Message: "Documents prepared for transfer",
		Data: endpoints.TransferData{
			IDCitizen:      idCitizen,
			TotalDocuments: len(transferDocuments),
			Documents:      transferDocuments,
			ExpiresIn:      "15m",
		},
	})
}
