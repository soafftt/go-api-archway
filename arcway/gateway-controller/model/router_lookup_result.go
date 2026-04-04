package model

import rewriteDTO "gateway/common/model/rewrite"

type RouterLookupError struct {
	Message string
	Detail  error
}

type RouterLookupResult struct {
	Ok          bool
	RewritePath rewriteDTO.RewritePathDTO
	Error       RouterLookupError
}

func NewRouterLookupResult(rewritePath rewriteDTO.RewritePathDTO) RouterLookupResult {
	return RouterLookupResult{
		Ok:          true,
		RewritePath: rewritePath,
		Error:       RouterLookupError{},
	}
}

func NewRouterLookupError(message string, detail error) RouterLookupResult {
	return RouterLookupResult{
		Ok:          false,
		RewritePath: rewriteDTO.NewEmptyRewritePathDTO(),
		Error:       RouterLookupError{Message: message, Detail: detail},
	}
}
