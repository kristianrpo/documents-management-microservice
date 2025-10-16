package util

// PaginationParams contains the normalized pagination parameters
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// NormalizePagination validates and normalizes pagination parameters
// Ensures page >= 1, limit is between 1 and 100, and calculates the offset
func NormalizePagination(page, limit int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
