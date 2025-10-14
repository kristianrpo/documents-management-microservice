package usecases

import (
	"context"
	"mime/multipart"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type DocumentService interface {
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerEmail string) (*domain.Document, error)
}

type documentService struct {
	repository   interfaces.DocumentRepository
	storage      interfaces.ObjectStorage
	hasher       util.FileHasher
	mimeDetector util.MimeTypeDetector
}

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

func (service *documentService) Upload(ctx context.Context, fileHeader *multipart.FileHeader, ownerEmail string) (*domain.Document, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, domain.NewFileReadError(err)
	}
	defer file.Close()

	hash, err := service.hasher.CalculateHash(file)
	if err != nil {
		return nil, domain.NewHashCalculateError(err)
	}

	objectKey := util.ObjectKeyFromHash(hash, fileHeader.Filename)

	contentType := service.mimeDetector.DetectFromFilename(fileHeader.Filename)

	if seeker, ok := file.(interface{ Seek(int64, int) (int64, error) }); ok {
		if _, err := seeker.Seek(0, 0); err != nil {
			return nil, domain.NewFileReadError(err)
		}
	}
	
	if err := service.storage.Put(ctx, file, objectKey, contentType); err != nil {
		return nil, domain.NewStorageUploadError(err)
	}

	publicURL := service.storage.PublicURL(objectKey)
	document := &domain.Document{
		Filename:   fileHeader.Filename,
		MimeType:   contentType,
		SizeBytes:  fileHeader.Size,
		HashSHA256: hash,
		Bucket:     service.storage.Bucket(),
		ObjectKey:  objectKey,
		URL:        publicURL,
		OwnerEmail: ownerEmail,
	}

	if err := document.Validate(); err != nil {
		return nil, err
	}

	if err := service.repository.Create(document); err != nil {
		return nil, domain.NewPersistenceError(err)
	}

	return document, nil
}
