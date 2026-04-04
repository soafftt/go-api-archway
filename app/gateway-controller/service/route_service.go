package service

import (
	"errors"

	"gateway/controller/component"

	code "gateway/common/code"
	"gateway/common/gjwt"
	rewriteDTO "gateway/common/model/rewrite"
	model "gateway/controller/model"
	modelDto "gateway/controller/model/dto"

	"github.com/google/wire"
)

type RouteService interface {
	GetRouteInfo(urlParseDto modelDto.URLParseDTO) model.RouterLookupResult
}

type routeService struct {
	routeCache component.RouteCache
}

func NewPolicyService(routeCache component.RouteCache) *routeService {
	return &routeService{routeCache: routeCache}
}

func (p *routeService) GetRouteInfo(urlParseDto modelDto.URLParseDTO) model.RouterLookupResult {
	upstreamService, ok := p.routeCache.Get(urlParseDto.Service)
	if !ok {
		return model.NewRouterLookupError(
			code.NOT_FOUND_UPSTREAM_SERVICE,
			errors.New("No matching upstream service found for: "+urlParseDto.String()),
		)
	}

	// 서브도메인이 있는 경우를 찾는다.
	domain, emptyDomain := upstreamService.LookupResourceDomain(urlParseDto.Domain)
	if domain == nil {
		return model.NewRouterLookupError(
			code.NOT_FOUND_UPSTREAM_DOMAIN,
			errors.New("No matching domain found for: "+urlParseDto.String()),
		)
	}

	// URI 경로를 찾는다.
	lookupPath := urlParseDto.GetPath(emptyDomain)
	pathStream := domain.LookupPath(lookupPath)
	if pathStream == nil {
		return model.NewRouterLookupError(
			code.NOT_FOUND_UPSTREAM_PATH,
			errors.New("No matching path found for: "+urlParseDto.String()),
		)
	}

	if pathStream.CheckAuthorization {
		// jwt 파싱
		codec, err := gjwt.NewCodec(upstreamService.ServiceName)
		if err != nil {
			return model.NewRouterLookupError(
				code.NOT_FOUND_JWT_CODEC,
				errors.New("No matching codec for: "+upstreamService.ServiceName),
			)
		}

		decodeResult := codec.Parse("")
		if decodeResult.Err != nil {
			var message string = code.FAIL_AUTHORIZATION

			// jwt signature error
			if errors.Is(decodeResult.Err, gjwt.ErrJwtSigned) {
				message = code.FAIL_AUTHORIZATION
			}

			// jwt expire
			if errors.Is(decodeResult.Err, gjwt.ErrJwtExpire) {
				message = code.EXPIRE_JWT
			}

			return model.NewRouterLookupError(
				message,
				decodeResult.Err,
			)
		}

		// 수많은 데이터중 가장 중요한 user_id 를 꺼낸다.
		data, ok := decodeResult.Claims[upstreamService.Authorization.UserKey]
		if !ok {
			return model.NewRouterLookupError(
				code.FAIL_AUTHORIZATION,
				errors.New("NOT_FOUND_USER_KEY"),
			)
		}

		// 로그인 회원키가 있는 경우.
		return model.NewRouterLookupResult(rewriteDTO.NewRewritePathDTOWithUserKey(domain.Host, pathStream, data))
	}

	// 로그인 회원키가 없어도 되는 경우.
	return model.NewRouterLookupResult(rewriteDTO.NewRewritePathDTO(domain.Host, pathStream))
}

var RouteServiceSet = wire.NewSet(
	NewPolicyService,
	wire.Bind(new(RouteService), new(*routeService)),
)
