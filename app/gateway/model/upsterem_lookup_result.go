package model

import "gateway/common/model/rewrite"

type UpstreamLookupError struct {
	Message string
	Detail  error
}

type UpstreamLookupResult struct {
	Ok       bool
	Upstream *rewrite.RewitePathDTO
	Error    UpstreamLookupError
}

func NewUpstreamLookupResult(upstream *rewrite.RewitePathDTO) UpstreamLookupResult {
	return UpstreamLookupResult{
		Ok:       true,
		Upstream: upstream,
		Error:    UpstreamLookupError{},
	}
}

func NewUpstreamLookupError(message string, detail error) UpstreamLookupResult {
	return UpstreamLookupResult{
		Ok:       false,
		Upstream: nil,
		Error:    UpstreamLookupError{Message: message, Detail: detail},
	}
}
