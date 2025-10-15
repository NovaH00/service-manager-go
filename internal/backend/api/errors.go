package api

type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
	Details      string `json:"details,omitempty"`
}

func NewError(message, details string) ErrorResponse {
	return ErrorResponse{
		ErrorMessage: message,
		Details:      details,
	}
}
