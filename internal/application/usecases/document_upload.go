package usecases

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DocumentService defines the interface for document upload operations
type DocumentService interface {
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerID int64) (*models.Document, error)
}

type documentService struct {
	repository   interfaces.DocumentRepository
	storage      interfaces.ObjectStorage
	hasher       util.FileHasher
	mimeDetector util.MimeTypeDetector
}

// NewDocumentService creates a new document upload service
func NewDocumentService(
	repository interfaces.DocumentRepository,
	storage interfaces.ObjectStorage,
	hasher util.FileHasher,
	mimeDetector util.MimeTypeDetector,
) DocumentService {
	return &documentService{
		repository:   repository,
		storage:      storage,
		hasher:       hasher,
		mimeDetector: mimeDetector,
	}
}

// Upload uploads a document to storage and saves its metadata to the repository
// If a document with the same hash already exists for the owner, returns the existing document
func (service *documentService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerID int64) (*models.Document, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.NewFileReadError(err)
	}
	defer func() { _ = file.Close() }()

	// Delegate to UploadFromReader which contains the shared logic
	if seeker, ok := file.(io.ReadSeeker); ok {
		return service.UploadFromReader(ctx, seeker, fileHeader.Filename, fileHeader.Size, ownerID)
	}

	// If not a ReadSeeker, copy into a buffer
	// Use io.ReadAll as fallback (file should be small enough for upload use-cases)
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.NewFileReadError(err)
	}
	return service.UploadFromReader(ctx, bytes.NewReader(data), fileHeader.Filename, fileHeader.Size, ownerID)
}

// UploadFromReader uploads a document reading from an io.ReadSeeker. It implements DocumentUploader.
func (service *documentService) UploadFromReader(ctx context.Context, r io.ReadSeeker, filename string, size int64, ownerID int64) (*models.Document, error) {
	// Compute hash
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, errors.NewFileReadError(err)
	}
	hash, err := service.hasher.CalculateHash(r)
	if err != nil {
		return nil, errors.NewHashCalculateError(err)
	}

	existingDoc, _ := service.repository.FindByHashAndOwnerID(ctx, hash, ownerID)
	if existingDoc != nil {
		return existingDoc, nil
	}

	objectKey := util.ObjectKeyFromHash(hash, filename)
	contentType := service.mimeDetector.DetectFromFilename(filename)

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, errors.NewFileReadError(err)
	}

	if err := service.storage.Put(ctx, r, objectKey, contentType); err != nil {
		return nil, errors.NewStorageUploadError(err)
	}

	publicURL := service.storage.PublicURL(objectKey)
	document := &models.Document{
		Filename:             filename,
		MimeType:             contentType,
		SizeBytes:            size,
		HashSHA256:           hash,
		Bucket:               service.storage.Bucket(),
		ObjectKey:            objectKey,
		URL:                  publicURL,
		OwnerID:              ownerID,
		AuthenticationStatus: models.AuthenticationStatusUnauthenticated,
	}

	if err := document.Validate(); err != nil {
		return nil, err
	}

	if err := service.repository.Create(ctx, document); err != nil {
		return nil, errors.NewPersistenceError(err)
	}

	return document, nil
}
