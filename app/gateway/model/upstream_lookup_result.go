package model

import "gateway/common/model/rewrite"

type LookupErrorKind string

const (
	LookupErrorTransport       LookupErrorKind = "transport"
	LookupErrorReadBody        LookupErrorKind = "read_body"
	LookupErrorDecodeErrorBody LookupErrorKind = "decode_error_body"
	LookupErrorDecodeBody      LookupErrorKind = "decode_body"
	LookupErrorUpstreamResult  LookupErrorKind = "upstream_result"
)

type UpstreamLookupError struct {
	Kind    LookupErrorKind
	Message string
	Detail  error
}

type UpstreamLookupResult struct {
	Ok       bool
	Upstream *rewrite.RewritePathDTO
	Error    UpstreamLookupError
}

func NewUpstreamLookupResult(upstream *rewrite.RewritePathDTO) UpstreamLookupResult {
	return UpstreamLookupResult{
		Ok:       true,
		Upstream: upstream,
		Error:    UpstreamLookupError{},
	}
}

func NewUpstreamLookupError(kind LookupErrorKind, message string, detail error) UpstreamLookupResult {
	return UpstreamLookupResult{
		Ok:       false,
		Upstream: nil,
		Error:    UpstreamLookupError{Kind: kind, Message: message, Detail: detail},
	}
}
