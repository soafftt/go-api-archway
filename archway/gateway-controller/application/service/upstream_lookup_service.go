package service

import (
	"core/errs"
	"core/gjwt"
	"errors"
	"fmt"
	"gateway/controller/application/port/in"
	"gateway/controller/application/port/in/dto"
	"gateway/controller/application/port/out/cache"
)

type UpstreamLookupService struct {
	routeCache cache.RouteCache
}

func NewUpstreamLookupService(routeCache cache.RouteCache) in.UpstreamLookupUseCase {
	return &UpstreamLookupService{
		routeCache: routeCache,
	}
}

/*
요청 dto 의 정보를 이용하여, cache 된 rewrite 를 위한 정보를 조회한다.
*/
func (u UpstreamLookupService) LookUp(lookupRequest dto.UpStreamLookupRequest) dto.UpStreamLookupResult {
	service, ok := u.routeCache.Get(lookupRequest.Service)
	if !ok {
		return dto.NewErrUpStreamLookupResult(errs.ERR_NOT_FOUND_SERVICE)
	}

	resource, isEmptyDomain := service.LookupResourceDomain(lookupRequest.Domain)
	if resource == nil {
		return dto.NewErrUpStreamLookupResult(errs.ERR_NOT_FOUND_DOMAIN_RESOURCE)
	}

	resourcePathMatch := resource.LookupPath(lookupRequest.GetRelativePath(isEmptyDomain))
	if resourcePathMatch == nil {
		return dto.NewErrUpStreamLookupResult(errs.ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH)
	}
	resourcePath := resourcePathMatch.Path

	var userKey any
	if resourcePath.CheckAuthorization {
		claimKey := "user_id"
		if service.Authorization != nil && service.Authorization.UserKey != "" {
			claimKey = service.Authorization.UserKey
		}

		id, err := getUserIdAndVerifyAccessToken(
			lookupRequest.Service,
			claimKey,
			lookupRequest.AccessToken,
		)

		if err != nil {
			return dto.NewErrUpStreamLookupResult(errors.Join(errs.ERR_JWT_DESERIALIZE, err))
		}

		userKey = id
	}

	info := dto.NewUpStreamInfo(
		resource.Host,
		resourcePathMatch.RewrittenPath,
		resourcePathMatch.OriginalPath,
		resourcePath.Method,
		resourcePath.RequestTimeout,
		resourcePath.ResponseTimeout,
		resourcePath.CacheTimeout,
		userKey,
		resourcePath.RateLimitCount,
	)
	return dto.NewUpStreamLookupResult(info)
}

/*
JWT 인증이 필요한경우 JWT 파싱을 시도한다.
*/
func getUserIdAndVerifyAccessToken(keyService string, claimKey string, accessToken *string) (any, error) {
	if accessToken == nil || *accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	codec, err := gjwt.NewCodec(keyService)
	if err != nil {
		return nil, err
	}

	// 함수를 타고 다니는 과정에서 복사되는 accessToken 은 꽤 데이터가 크기에
	// 포인터로 이동시키다가 마지막에 복사하여 사용.
	parseResult := codec.Parse(*accessToken)
	if parseResult.Err != nil {
		return nil, parseResult.Err
	}

	if !parseResult.Valid {
		return nil, fmt.Errorf("jwt token is invalid: %v", errors.New(gjwt.ErrJwtSigned.Error()))
	}

	claim, has := parseResult.Claims[claimKey]
	if !has {
		return nil, fmt.Errorf("jwt claim %s missing", claimKey)
	}

	claimValue, ok := claim.(string)
	if !ok || claimValue == "" {
		return nil, fmt.Errorf("jwt claim %s must be a non-empty string", claimKey)
	}

	return claimValue, nil
}
