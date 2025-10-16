package util_test

import (
	"testing"

	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	"github.com/stretchr/testify/assert"
)

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedPage   int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "valid pagination",
			page:           2,
			limit:          20,
			expectedPage:   2,
			expectedLimit:  20,
			expectedOffset: 20,
		},
		{
			name:           "page less than 1 defaults to 1",
			page:           0,
			limit:          10,
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "negative page defaults to 1",
			page:           -5,
			limit:          10,
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "limit less than 1 defaults to 10",
			page:           1,
			limit:          0,
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "negative limit defaults to 10",
			page:           1,
			limit:          -20,
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "limit greater than 100 caps at 100",
			page:           1,
			limit:          150,
			expectedPage:   1,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "limit exactly 100",
			page:           2,
			limit:          100,
			expectedPage:   2,
			expectedLimit:  100,
			expectedOffset: 100,
		},
		{
			name:           "limit exactly 1",
			page:           3,
			limit:          1,
			expectedPage:   3,
			expectedLimit:  1,
			expectedOffset: 2,
		},
		{
			name:           "page 1 has offset 0",
			page:           1,
			limit:          25,
			expectedPage:   1,
			expectedLimit:  25,
			expectedOffset: 0,
		},
		{
			name:           "page 5 with limit 10",
			page:           5,
			limit:          10,
			expectedPage:   5,
			expectedLimit:  10,
			expectedOffset: 40,
		},
		{
			name:           "both invalid values",
			page:           -1,
			limit:          -1,
			expectedPage:   1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "page and limit both exceed bounds",
			page:           0,
			limit:          200,
			expectedPage:   1,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "large page number",
			page:           1000,
			limit:          50,
			expectedPage:   1000,
			expectedLimit:  50,
			expectedOffset: 49950,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.NormalizePagination(tt.page, tt.limit)

			assert.Equal(t, tt.expectedPage, result.Page, "Page mismatch")
			assert.Equal(t, tt.expectedLimit, result.Limit, "Limit mismatch")
			assert.Equal(t, tt.expectedOffset, result.Offset, "Offset mismatch")
		})
	}
}
