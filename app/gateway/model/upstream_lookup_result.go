package model

import (
	"gateway/common/model/rewrite"
	"gateway/gatewayerrors"
)

type UpstreamLookupError struct {
	Kind    gatewayerrors.LookupErrorKind
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

func NewUpstreamLookupError(kind gatewayerrors.LookupErrorKind, message string, detail error) UpstreamLookupResult {
	return UpstreamLookupResult{
		Ok:       false,
		Upstream: nil,
		Error:    UpstreamLookupError{Kind: kind, Message: message, Detail: detail},
	}
}
