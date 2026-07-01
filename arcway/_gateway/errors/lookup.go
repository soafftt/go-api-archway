package gwerrors

// LookupErrorKind upstream lookup 에러 분류
type LookupErrorKind string

const (
	LookupErrorTransport       LookupErrorKind = "transport"
	LookupErrorReadBody        LookupErrorKind = "read_body"
	LookupErrorDecodeErrorBody LookupErrorKind = "decode_error_body"
	LookupErrorDecodeBody      LookupErrorKind = "decode_body"
	LookupErrorUpstreamResult  LookupErrorKind = "upstream_result"
)

// upstream lookup 서비스 에러 메시지 코드
const (
	ErrMsgTransport       = "UNIX_SOCKET_RESPONSE_ERROR"
	ErrMsgReadBody        = "UNIX_SOCKET_RESPONSE_BODY_READ_ERROR"
	ErrMsgDecodeBody      = "UNIX_SOCKET_RESPONSE_BODY_JSON_UNMARSHAL_ERROR"
	ErrMsgDecodeErrorBody = "UNIX_SOCKET_RESPONSE_ERROR_BODY_JSON_UNMARSHAL_ERROR"
)
