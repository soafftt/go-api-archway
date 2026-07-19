package in

import (
	"core/errs"
	"core/gjwt"
	"errors"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func NewErrorResponse(err error) ErrorResponse {
	archwayError, ok := errors.AsType[errs.ArchwayError](err)
	if ok {
		return ErrorResponse{
			Message: archwayError.Error(),
			Detail:  err.Error(),
		}
	}

	jwtError, ok := errors.AsType[gjwt.JwtError](err)
	if ok {
		return ErrorResponse{
			Message: jwtError.Error(),
			Detail:  err.Error(),
		}
	}

	return ErrorResponse{
		Message: errs.ERR_UNKNOWN.Error(),
		Detail:  err.Error(),
	}
}
