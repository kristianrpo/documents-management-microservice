package request_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/request"
	"github.com/stretchr/testify/assert"
)

func TestListDocumentsRequest(t *testing.T) {
	t.Run("valid request with all fields", func(t *testing.T) {
		req := request.ListDocumentsRequest{
			IDCitizen: 12345,
			Page:      1,
			Limit:     10,
		}

		assert.Equal(t, int64(12345), req.IDCitizen)
		assert.Equal(t, 1, req.Page)
		assert.Equal(t, 10, req.Limit)
	})

	t.Run("request with different page and limit", func(t *testing.T) {
		req := request.ListDocumentsRequest{
			IDCitizen: 67890,
			Page:      5,
			Limit:     50,
		}

		assert.Equal(t, int64(67890), req.IDCitizen)
		assert.Equal(t, 5, req.Page)
		assert.Equal(t, 50, req.Limit)
	})

	t.Run("request with minimum values", func(t *testing.T) {
		req := request.ListDocumentsRequest{
			IDCitizen: 1,
			Page:      1,
			Limit:     1,
		}

		assert.Equal(t, int64(1), req.IDCitizen)
		assert.Equal(t, 1, req.Page)
		assert.Equal(t, 1, req.Limit)
	})

	t.Run("request with maximum limit", func(t *testing.T) {
		req := request.ListDocumentsRequest{
			IDCitizen: 12345,
			Page:      1,
			Limit:     100,
		}

		assert.Equal(t, 100, req.Limit)
	})
}

func TestUploadRequest(t *testing.T) {
	t.Run("request structure", func(t *testing.T) {
		req := request.UploadRequest{
			File:      nil, // In real scenario, this would be a *multipart.FileHeader
			IDCitizen: 12345,
		}

		assert.Equal(t, int64(12345), req.IDCitizen)
		assert.Nil(t, req.File)
	})

	t.Run("request with different citizen IDs", func(t *testing.T) {
		citizenIDs := []int64{1, 12345, 67890, 999999}

		for _, id := range citizenIDs {
			req := request.UploadRequest{
				IDCitizen: id,
			}

			assert.Equal(t, id, req.IDCitizen)
		}
	})
}
