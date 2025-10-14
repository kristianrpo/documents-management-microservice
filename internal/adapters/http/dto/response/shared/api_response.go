package shared

type APIResponse[T any] struct {
	Success bool         `json:"success"`
	Data    *T           `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

func NewSuccessResponse[T any](data T) APIResponse[T] {
	return APIResponse[T]{
		Success: true,
		Data:    &data,
	}
}

func NewErrorResponse(code, message string) APIResponse[any] {
	return APIResponse[any]{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
