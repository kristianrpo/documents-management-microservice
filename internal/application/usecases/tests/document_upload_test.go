package usecases

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDocumentUploadService_Execute_Success(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	fileContent := []byte("test content")
	file := newMultipartFileHeader("test.pdf", fileContent)
	hash := "a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9" // 64 hex chars
	mimeType := "application/pdf"
	bucketName := "test-bucket"
	publicURL := "https://s3.amazonaws.com/test-bucket/documents/test.pdf"

	hasher.On("CalculateHash", mock.Anything).Return(hash, nil)
	mimeDetector.On("DetectFromFilename", "test.pdf").Return(mimeType)
	repo.On("FindByHashAndOwnerID", ctx, hash, ownerID).Return(nil, nil)
	storage.On("Bucket").Return(bucketName)
	storage.On("Put", ctx, mock.Anything, mock.AnythingOfType("string"), mimeType).Return(nil)
	storage.On("PublicURL", mock.AnythingOfType("string")).Return(publicURL)
	repo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ownerID, result.OwnerID)
	assert.Equal(t, "test.pdf", result.Filename)
	assert.Equal(t, mimeType, result.MimeType)
	assert.Equal(t, publicURL, result.URL)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
	hasher.AssertExpectations(t)
	mimeDetector.AssertExpectations(t)
}

func TestDocumentUploadService_Execute_HashError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	file := newMultipartFileHeader("test.pdf", []byte("test content"))
	expectedError := errors.New("hash computation failed")
	hasher.On("CalculateHash", mock.Anything).Return("", expectedError)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to calculate file hash")

	hasher.AssertExpectations(t)
}

func TestDocumentUploadService_Execute_DefaultMimeTypeWhenUnknown(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	file := newMultipartFileHeader("test.unknown", []byte("test content"))
	hash := "a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9"
	hasher.On("CalculateHash", mock.Anything).Return(hash, nil)
	mimeDetector.On("DetectFromFilename", "test.unknown").Return("application/octet-stream")
	repo.On("FindByHashAndOwnerID", ctx, hash, ownerID).Return(nil, nil)
	storage.On("Bucket").Return("bucket")
	storage.On("Put", ctx, mock.Anything, mock.AnythingOfType("string"), "application/octet-stream").Return(nil)
	storage.On("PublicURL", mock.AnythingOfType("string")).Return("https://example.com/object")
	repo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "application/octet-stream", result.MimeType)

	hasher.AssertExpectations(t)
	mimeDetector.AssertExpectations(t)
}

func TestDocumentUploadService_Execute_DuplicateDocument(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	file := newMultipartFileHeader("test.pdf", []byte("test content"))
	hash := "abcd1234" // not used for validation in duplicate path

	existingDoc := &models.Document{ID: "existing-id", Filename: "existing.pdf", OwnerID: ownerID}
	hasher.On("CalculateHash", mock.Anything).Return(hash, nil)
	// mime detection is not called because duplicate is returned before
	repo.On("FindByHashAndOwnerID", ctx, hash, ownerID).Return(existingDoc, nil)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingDoc, result)

	repo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	mimeDetector.AssertExpectations(t)
}

func TestDocumentUploadService_Execute_StorageError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	file := newMultipartFileHeader("test.pdf", []byte("test content"))
	hash := "abcd1234"

	hasher.On("CalculateHash", mock.Anything).Return(hash, nil)
	mimeDetector.On("DetectFromFilename", "test.pdf").Return("application/pdf")
	repo.On("FindByHashAndOwnerID", ctx, hash, ownerID).Return(nil, nil)

	expectedError := errors.New("storage error")
	storage.On("Put", ctx, mock.Anything, mock.AnythingOfType("string"), "application/pdf").Return(expectedError)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to upload to storage")

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
	hasher.AssertExpectations(t)
	mimeDetector.AssertExpectations(t)
}

func TestDocumentUploadService_Execute_RepositoryCreateError(t *testing.T) {
	// Arrange
	repo := new(MockDocumentRepository)
	storage := new(MockObjectStorage)
	hasher := new(MockFileHasher)
	mimeDetector := new(MockMimeDetector)

	service := usecases.NewDocumentService(repo, storage, hasher, mimeDetector)

	ctx := context.Background()
	ownerID := int64(1)
	file := newMultipartFileHeader("test.pdf", []byte("test content"))
	hash := "a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9a3f1b0f9"

	hasher.On("CalculateHash", mock.Anything).Return(hash, nil)
	mimeDetector.On("DetectFromFilename", "test.pdf").Return("application/pdf")
	repo.On("FindByHashAndOwnerID", ctx, hash, ownerID).Return(nil, nil)
	storage.On("Bucket").Return("test-bucket")
	storage.On("Put", ctx, mock.Anything, mock.AnythingOfType("string"), "application/pdf").Return(nil)
	storage.On("PublicURL", mock.AnythingOfType("string")).Return("https://s3.amazonaws.com/test/doc.pdf")

	expectedError := errors.New("database error")
	repo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(expectedError)

	// Act
	result, err := service.Upload(ctx, file, ownerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to persist document")

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
	hasher.AssertExpectations(t)
	mimeDetector.AssertExpectations(t)
}

// helper to build a multipart.FileHeader with content
func newMultipartFileHeader(filename string, content []byte) *multipart.FileHeader {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	fw, _ := w.CreateFormFile("file", filename)
	_, _ = fw.Write(content)
	_ = w.Close()

	r := multipart.NewReader(bytes.NewReader(b.Bytes()), w.Boundary())
	form, _ := r.ReadForm(int64(len(content) + 1024))
	return form.File["file"][0]
}
