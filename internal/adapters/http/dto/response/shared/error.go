package shared

type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"invalid request format or validation failed"`
}

func NewErrorDetail(code, message string) ErrorDetail {
	return ErrorDetail{
		Code:    code,
		Message: message,
	}
}
