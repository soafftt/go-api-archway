package model

import rewiterDto "gateway/common/model/rewrite"

type RouterLookupError struct {
	Message string
	Detail  error
}

type RouterLookupResult struct {
	Ok         bool
	RewitePath rewiterDto.RewitePathDTO
	Error      RouterLookupError
}

func NewRouterLookupResult(rewritePath rewiterDto.RewitePathDTO) RouterLookupResult {
	return RouterLookupResult{
		Ok:         true,
		RewitePath: rewritePath,
		Error:      RouterLookupError{},
	}
}

func NewRoterLookupError(message string, detail error) RouterLookupResult {
	return RouterLookupResult{
		Ok:         false,
		RewitePath: rewiterDto.NewEmptyRewitePathDTO(),
		Error:      RouterLookupError{Message: message, Detail: detail},
	}
}
