package service

import (
	"gateway/application/port/in"
	"gateway/application/port/out"
)

type upstreamLookupService struct {
	httpUpstreamLookupPort out.UpstreamLookupPort
	grpcUpstreamLookupPort out.UpstreamLookupGrpcPort
}

func NewUpstreamLookupService(
	httpUpstreamLookupPort out.UpstreamLookupPort,
	grpcUpstreamLookupPort out.UpstreamLookupGrpcPort,

) in.UpstreamLookupUseCase {
	return &upstreamLookupService{
		httpUpstreamLookupPort: httpUpstreamLookupPort,
		grpcUpstreamLookupPort: grpcUpstreamLookupPort,
	}
}

func (u *upstreamLookupService) Lookup(srcPath string, accessToken *string, transport in.Transport) (in.UpstreamLookupResult, error) {
	if transport == in.UnixHttp {
		return u.lookupByHttp(srcPath, accessToken)
	}

	return u.lookupByGrpc(srcPath, accessToken)
}

func (u *upstreamLookupService) lookupByHttp(srcPath string, accessToken *string) (in.UpstreamLookupResult, error) {
	var result in.UpstreamLookupResult

	lookupResult, err := u.grpcUpstreamLookupPort.GetUpstreamInfo(srcPath, accessToken)
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

func (u *upstreamLookupService) lookupByGrpc(srcPath string, accessToken *string) (in.UpstreamLookupResult, error) {
	var result in.UpstreamLookupResult
	lookupResult, err := u.grpcUpstreamLookupPort.GetUpstreamInfo(srcPath, accessToken)
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
