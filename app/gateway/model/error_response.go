package model

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

func NewErrorResponse(message string, details string) ErrorResponse {
	return ErrorResponse{Message: message, Details: details}
}
