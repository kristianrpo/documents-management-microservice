package interfaces

import (
	"context"
	"io"

	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentUploader defines an interface for uploading from an io.ReadSeeker.
// This is implemented by the document upload usecase so other packages (e.g., event handlers)
// can reuse the same upload logic without depending on the concrete implementation.
type DocumentUploader interface {
	UploadFromReader(ctx context.Context, r io.ReadSeeker, filename string, size int64, ownerID int64) (*models.Document, error)
}
