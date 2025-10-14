package request

import "mime/multipart"

type UploadRequest struct {
	File  *multipart.FileHeader `form:"file" binding:"required"`
	Email string                `form:"email" binding:"required,email"`
}
