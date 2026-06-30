package errs

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
)
