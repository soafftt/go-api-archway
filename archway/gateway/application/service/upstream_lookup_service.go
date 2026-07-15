package service

import (
	"gateway/application/port/in"
	"gateway/application/port/out"
)

type upstreamLookupService struct {
	upstreamLookupPort out.UpstreamLookupPort
}

func NewUpstreamLookupService(
	upstreamLookupPort out.UpstreamLookupPort,

) in.UpstreamLookupUseCase {
	return &upstreamLookupService{
		upstreamLookupPort: upstreamLookupPort,
	}
}

func (u upstreamLookupService) Lookup(srcPath string, accessToken *string) (in.UpstreamLookupResult, error) {
	var result in.UpstreamLookupResult

	lookupResult, err := u.upstreamLookupPort.GetUpstreamInfo(srcPath, accessToken)
	if err != nil {
		return result, err
	}

	result = in.UpstreamLookupResult{
		ServiceName:     lookupResult.ServiceName,
		Host:            lookupResult.Host,
		Path:            lookupResult.Path,
		OriginPath:      lookupResult.OriginalPath,
		Method:          lookupResult.Method,
		ResponseTimeout: lookupResult.ResponseTimeout,
		RequestTimeout:  lookupResult.RequestTimeout,
		CacheTimeout:    lookupResult.CacheTimeout,
		UserKey:         lookupResult.UserKey,
		RateLimitCount:  lookupResult.RateLimitCount,
	}

	return result, nil
}
