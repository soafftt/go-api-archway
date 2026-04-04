package gatewayerrors

// LookupErrorKind upstream lookup 에러 분류
type LookupErrorKind string

const (
	ErrLookupTransport       LookupErrorKind = "transport"
	ErrLookupReadBody        LookupErrorKind = "read_body"
	ErrLookupDecodeErrorBody LookupErrorKind = "decode_error_body"
	ErrLookupDecodeBody      LookupErrorKind = "decode_body"
	ErrLookupUpstreamResult  LookupErrorKind = "upstream_result"
)

// upstream lookup 서비스 에러 메시지 코드
const (
	ErrMsgTransport       = "UNIX_SOCKET_RESPONSE_ERROR"
	ErrMsgReadBody        = "UNIX_SOCKET_RESPONSE_BODY_READ_ERROR"
	ErrMsgDecodeBody      = "UNIX_SOCKET_RESPONSE_BODY_JSON_UNMARSHAL_ERROR"
	ErrMsgDecodeErrorBody = "UNIX_SOCKET_RESPONSE_ERROR_BODY_JSON_UNMARSHAL_ERROR"
)
