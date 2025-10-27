package usecases

import (
	"context"
	"io"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/mock"
)

// MockDocumentRepository is a mock implementation of DocumentRepository
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) FindByHashAndOwnerID(ctx context.Context, hashSHA256 string, ownerID int64) (*models.Document, error) {
	args := m.Called(ctx, hashSHA256, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) List(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, int64, error) {
	args := m.Called(ctx, ownerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Document), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) DeleteByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error) {
	args := m.Called(ctx, ownerID)
	return args.Int(0), args.Error(1)
}

func (m *MockDocumentRepository) UpdateAuthenticationStatus(ctx context.Context, documentID string, status models.AuthenticationStatus) error {
	args := m.Called(ctx, documentID, status)
	return args.Error(0)
}

func (m *MockDocumentRepository) EnsureTableExists(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockObjectStorage is a mock implementation of ObjectStorage
type MockObjectStorage struct {
	mock.Mock
}

func (m *MockObjectStorage) Put(ctx context.Context, body io.Reader, objectKey, contentType string) error {
	args := m.Called(ctx, body, objectKey, contentType)
	return args.Error(0)
}

func (m *MockObjectStorage) PublicURL(objectKey string) string {
	args := m.Called(objectKey)
	return args.String(0)
}

func (m *MockObjectStorage) Bucket() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockObjectStorage) Delete(ctx context.Context, objectKey string) error {
	args := m.Called(ctx, objectKey)
	return args.Error(0)
}

func (m *MockObjectStorage) GeneratePresignedURL(ctx context.Context, objectKey string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiration)
	return args.String(0), args.Error(1)
}

// MockFileHasher is a mock implementation of FileHasher
type MockFileHasher struct {
	mock.Mock
}

func (m *MockFileHasher) CalculateHash(reader io.Reader) (string, error) {
	args := m.Called(reader)
	return args.String(0), args.Error(1)
}

// MockMimeDetector is a mock implementation of MimeDetector
type MockMimeDetector struct {
	mock.Mock
}

func (m *MockMimeDetector) DetectFromFilename(filename string) string {
	args := m.Called(filename)
	return args.String(0)
}

// MockMessagePublisher is a mock implementation of MessagePublisher
type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) Publish(ctx context.Context, queue string, message []byte) error {
	args := m.Called(ctx, queue, message)
	return args.Error(0)
}

func (m *MockMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}
