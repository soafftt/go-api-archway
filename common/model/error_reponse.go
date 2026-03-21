package model

type ErroeResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}
