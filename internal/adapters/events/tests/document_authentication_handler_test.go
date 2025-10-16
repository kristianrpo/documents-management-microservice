package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	adapters "github.com/kristianrpo/document-management-microservice/internal/adapters/events"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, _ *models.Document) error { return nil }
func (m *mockRepo) FindByHashAndOwnerID(ctx context.Context, hash string, ownerID int64) (*models.Document, error) {
	args := m.Called(ctx, hash, ownerID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Document), args.Error(1)
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Document), args.Error(1)
}
func (m *mockRepo) List(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, int64, error) {
	args := m.Called(ctx, ownerID, limit, offset)
	if args.Get(0) == nil { return nil, 0, args.Error(2) }
	return args.Get(0).([]*models.Document), int64(args.Int(1)), args.Error(2)
}
func (m *mockRepo) DeleteByID(ctx context.Context, id string) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*models.Document), args.Error(1)
}
func (m *mockRepo) DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error) {
	args := m.Called(ctx, ownerID)
	return args.Int(0), args.Error(1)
}
func (m *mockRepo) UpdateAuthenticationStatus(ctx context.Context, documentID string, status models.AuthenticationStatus) error {
	args := m.Called(ctx, documentID, status)
	return args.Error(0)
}

func TestHandleAuthenticationCompleted_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	h := adapters.NewDocumentAuthenticationHandler(repo)

	evt := events.DocumentAuthenticationCompletedEvent{
		DocumentID: "doc-1",
		IDCitizen:  99,
		Authenticated: true,
		Message:   "ok",
	}
	payload, _ := json.Marshal(evt)
	repo.On("UpdateAuthenticationStatus", ctx, "doc-1", models.AuthenticationStatusAuthenticated).Return(nil)

	err := h.HandleAuthenticationCompleted(ctx, payload)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandleAuthenticationCompleted_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	h := adapters.NewDocumentAuthenticationHandler(repo)
	// invalid JSON
	payload := []byte("{invalid}")
	err := h.HandleAuthenticationCompleted(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestHandleAuthenticationCompleted_UpdateError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	h := adapters.NewDocumentAuthenticationHandler(repo)
	evt := events.DocumentAuthenticationCompletedEvent{DocumentID: "doc-1", IDCitizen: 3, Authenticated: false}
	payload, _ := json.Marshal(evt)
	repo.On("UpdateAuthenticationStatus", ctx, "doc-1", models.AuthenticationStatusUnauthenticated).Return(errors.New("db err"))

	err := h.HandleAuthenticationCompleted(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update")
	repo.AssertExpectations(t)
}
