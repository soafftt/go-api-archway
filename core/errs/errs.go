package errs

import "errors"

type ArchwayError string

func (e ArchwayError) Error() string {
	return string(e)
}

const (
	ERR_UNKNOWN                        ArchwayError = "unknown"
	ERR_NOT_FOUND_SERVICE              ArchwayError = "ERR_NOT_FOUND_SERVICE"
	ERR_NOT_FOUND_DOMAIN_RESOURCE      ArchwayError = "ERR_NOT_FOUND_DOMAIN_RESOURCE"
	ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH ArchwayError = "ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH"
	ERR_JWT_DESERIALIZE                ArchwayError = "ERR_JWT_DESERIALIZE"
	ERR_INVALID_TARGET                 ArchwayError = "ERR_INVALID_TARGET"
	ERR_GATEWAY_CONTROLLER_SEND        ArchwayError = "ERR_GATEWAY_CONTROLLER_SEND"
)

// 검색효율을 위하여 공간을 희생함.
var archwayErrorMap = map[string]ArchwayError{
	ERR_UNKNOWN.Error():                        ERR_UNKNOWN,
	ERR_NOT_FOUND_SERVICE.Error():              ERR_NOT_FOUND_SERVICE,
	ERR_NOT_FOUND_DOMAIN_RESOURCE.Error():      ERR_NOT_FOUND_DOMAIN_RESOURCE,
	ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH.Error(): ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH,
	ERR_JWT_DESERIALIZE.Error():                ERR_JWT_DESERIALIZE,
	ERR_INVALID_TARGET.Error():                 ERR_INVALID_TARGET,
}

func ToArchwayError(value string) error {
	archwayError, ok := archwayErrorMap[value]
	if !ok {
		return errors.Join(ERR_UNKNOWN, errors.New(value))
	}

	return archwayError
}

func ToArchwayFromError(err error) ArchwayError {
	archwayError, ok := archwayErrorMap[err.Error()]
	if !ok {
		return ERR_UNKNOWN
	}

	return archwayError
}
