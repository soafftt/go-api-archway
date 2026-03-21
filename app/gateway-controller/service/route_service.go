package service

import (
	"errors"
	"gateway/controller/component"

	code "gateway/common/code"
	rewiterDto "gateway/common/model/rewrite"
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
		return model.NewRoterLookupError(
			code.NOT_FOUND_UPSTERAM_SERVICE,
			errors.New("No matching upstream service found for: "+urlParseDto.String()),
		)
	}

	// 서브도메인이 있는 경우를 찾는다.
	domain, emptyDomain := upstreamService.LookupResourceDomain(urlParseDto.Domain)
	if domain == nil {
		return model.NewRoterLookupError(
			code.NOT_FOUND_UPSTERAM_SERVICE,
			errors.New("No matching domain found for: "+urlParseDto.String()),
		)
	}

	// URI 경로를 찾는다.
	lookupPath := urlParseDto.GetPath(emptyDomain)
	pathStream := domain.LookupPath(lookupPath)
	if pathStream == nil {
		return model.NewRoterLookupError(
			code.NOT_FOUND_UPSTERAM_SERVICE,
			errors.New("No matching path found for: "+urlParseDto.String()),
		)
	}

	return model.NewRouterLookupResult(rewiterDto.NewRewitePathDTO(domain.Host, pathStream))
}

var RouteServiceSet = wire.NewSet(
	NewPolicyService,
	wire.Bind(new(RouteService), new(*routeService)),
)
