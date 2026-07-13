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

	controllerLookupResult, err := u.upstreamLookupPort.GetUpstreamInfo(srcPath, accessToken)
	if err != nil {
		return result, err
	}

	result = in.UpstreamLookupResult{
		Host:            controllerLookupResult.Host,
		Path:            controllerLookupResult.Path,
		OriginPath:      controllerLookupResult.OriginalPath,
		Method:          controllerLookupResult.Method,
		ResponseTimeout: controllerLookupResult.ResponseTimeout,
		RequestTimeout:  controllerLookupResult.RequestTimeout,
		CacheTimeout:    controllerLookupResult.CacheTimeout,
		UserKey:         controllerLookupResult.UserKey,
	}

	return result, nil
}
