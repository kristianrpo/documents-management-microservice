package shared_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
	"github.com/stretchr/testify/assert"
)

func TestPagination_Structure(t *testing.T) {
	t.Run("pagination with valid values", func(t *testing.T) {
		pagination := shared.Pagination{
			Page:       1,
			Limit:      10,
			TotalItems: 100,
			TotalPages: 10,
		}

		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.Limit)
		assert.Equal(t, int64(100), pagination.TotalItems)
		assert.Equal(t, 10, pagination.TotalPages)
	})

	t.Run("pagination with first page", func(t *testing.T) {
		pagination := shared.Pagination{
			Page:       1,
			Limit:      25,
			TotalItems: 50,
			TotalPages: 2,
		}

		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 25, pagination.Limit)
		assert.Equal(t, int64(50), pagination.TotalItems)
		assert.Equal(t, 2, pagination.TotalPages)
	})

	t.Run("pagination with last page", func(t *testing.T) {
		pagination := shared.Pagination{
			Page:       5,
			Limit:      10,
			TotalItems: 42,
			TotalPages: 5,
		}

		assert.Equal(t, 5, pagination.Page)
		assert.Equal(t, 10, pagination.Limit)
		assert.Equal(t, int64(42), pagination.TotalItems)
		assert.Equal(t, 5, pagination.TotalPages)
	})

	t.Run("pagination with zero items", func(t *testing.T) {
		pagination := shared.Pagination{
			Page:       1,
			Limit:      10,
			TotalItems: 0,
			TotalPages: 0,
		}

		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.Limit)
		assert.Equal(t, int64(0), pagination.TotalItems)
		assert.Equal(t, 0, pagination.TotalPages)
	})

	t.Run("pagination with large dataset", func(t *testing.T) {
		pagination := shared.Pagination{
			Page:       100,
			Limit:      50,
			TotalItems: 10000,
			TotalPages: 200,
		}

		assert.Equal(t, 100, pagination.Page)
		assert.Equal(t, 50, pagination.Limit)
		assert.Equal(t, int64(10000), pagination.TotalItems)
		assert.Equal(t, 200, pagination.TotalPages)
	})
}
