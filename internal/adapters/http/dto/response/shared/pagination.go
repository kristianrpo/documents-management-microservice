package shared

type Pagination struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	TotalItems int64 `json:"total_items" example:"42"`
	TotalPages int   `json:"total_pages" example:"5"`
}
