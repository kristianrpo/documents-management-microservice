package request

type ListDocumentsRequest struct {
	IDCitizen int64 `form:"id_citizen" binding:"required,gt=0"`
	Page      int   `form:"page" binding:"min=1" example:"1"`
	Limit     int   `form:"limit" binding:"min=1,max=100" example:"10"`
}
