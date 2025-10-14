package request

type ListDocumentsRequest struct {
	Email string `form:"email" binding:"required,email"`
	Page  int    `form:"page" binding:"min=1" example:"1"`
	Limit int    `form:"limit" binding:"min=1,max=100" example:"10"`
}
