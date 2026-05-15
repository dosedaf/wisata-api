package utils

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func SuccessResponse(message string, data any) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(message string, errors any) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
