package request

import "mime/multipart"

type UploadRequest struct {
	File      *multipart.FileHeader `form:"file" binding:"required"`
	IDCitizen int64                 `form:"id_citizen" binding:"required,gt=0"`
}
