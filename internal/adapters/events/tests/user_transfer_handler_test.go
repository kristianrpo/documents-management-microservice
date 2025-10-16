package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	adapters "github.com/kristianrpo/document-management-microservice/internal/adapters/events"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDeleteAllService struct{ mock.Mock }

func (m *mockDeleteAllService) DeleteAll(ctx context.Context, ownerID int64) (int, error) {
	args := m.Called(ctx, ownerID)
	return args.Int(0), args.Error(1)
}

func TestHandleUserTransferred_Success(t *testing.T) {
	ctx := context.Background()
	service := new(mockDeleteAllService)
	h := adapters.NewUserTransferHandler(service)
	_ = usecases.DocumentDeleteAllService(nil) // type reference for import

	e := events.UserTransferredEvent{IDCitizen: 42}
	payload, _ := json.Marshal(e)
	service.On("DeleteAll", ctx, int64(42)).Return(7, nil)

	err := h.HandleUserTransferred(ctx, payload)
	assert.NoError(t, err)
	service.AssertExpectations(t)
}

func TestHandleUserTransferred_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	service := new(mockDeleteAllService)
	h := adapters.NewUserTransferHandler(service)
	payload := []byte("{invalid}")
	err := h.HandleUserTransferred(ctx, payload)
	assert.Error(t, err)
}

func TestHandleUserTransferred_DeleteError(t *testing.T) {
	ctx := context.Background()
	service := new(mockDeleteAllService)
	h := adapters.NewUserTransferHandler(service)
	e := events.UserTransferredEvent{IDCitizen: 42}
	payload, _ := json.Marshal(e)
	service.On("DeleteAll", ctx, int64(42)).Return(0, errors.New("boom"))

	err := h.HandleUserTransferred(ctx, payload)
	assert.Error(t, err)
	service.AssertExpectations(t)
}
