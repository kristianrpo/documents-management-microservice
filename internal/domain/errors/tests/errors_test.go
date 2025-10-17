package errors_test

import (
	"errors"
	"testing"

	domainerrors "github.com/kristianrpo/document-management-microservice/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestDomainError_Error(t *testing.T) {
	err := &domainerrors.DomainError{
		Code:    domainerrors.ErrCodeValidation,
		Message: "test error message",
	}

	assert.Equal(t, "test error message", err.Error())
}

func TestDomainError_Unwrap(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("underlying error")
		err := &domainerrors.DomainError{
			Code:    domainerrors.ErrCodeFileRead,
			Message: "failed to read",
			Err:     wrappedErr,
		}

		assert.Equal(t, wrappedErr, err.Unwrap())
	})

	t.Run("without wrapped error", func(t *testing.T) {
		err := &domainerrors.DomainError{
			Code:    domainerrors.ErrCodeValidation,
			Message: "validation failed",
		}

		assert.Nil(t, err.Unwrap())
	})
}

func TestNewValidationError(t *testing.T) {
	err := domainerrors.NewValidationError("invalid input")

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodeValidation, err.Code)
	assert.Equal(t, "invalid input", err.Message)
	assert.Nil(t, err.Err)
}

func TestNewFileReadError(t *testing.T) {
	underlyingErr := errors.New("file not found")
	err := domainerrors.NewFileReadError(underlyingErr)

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodeFileRead, err.Code)
	assert.Equal(t, "failed to read file", err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestNewHashCalculateError(t *testing.T) {
	underlyingErr := errors.New("hash computation failed")
	err := domainerrors.NewHashCalculateError(underlyingErr)

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodeHashCalculate, err.Code)
	assert.Equal(t, "failed to calculate file hash", err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestNewStorageUploadError(t *testing.T) {
	underlyingErr := errors.New("S3 upload failed")
	err := domainerrors.NewStorageUploadError(underlyingErr)

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodeStorageUpload, err.Code)
	assert.Equal(t, "failed to upload to storage", err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestNewPersistenceError(t *testing.T) {
	underlyingErr := errors.New("database connection failed")
	err := domainerrors.NewPersistenceError(underlyingErr)

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodePersistence, err.Code)
	assert.Equal(t, "failed to persist document", err.Message)
	assert.Equal(t, underlyingErr, err.Err)
}

func TestNewNotFoundError(t *testing.T) {
	err := domainerrors.NewNotFoundError("document not found")

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrCodeNotFound, err.Code)
	assert.Equal(t, "document not found", err.Message)
	assert.Nil(t, err.Err)
}
