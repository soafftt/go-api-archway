package model

type ErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func NewErrorResponse(message string, detail string) ErrorResponse {
	return ErrorResponse{Message: message, Detail: detail}
}
