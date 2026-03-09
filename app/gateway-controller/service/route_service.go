package service

import (
	"fmt"
	"gateway/controller/component"

	modelDto "gateway/controller/model/dto"

	"github.com/google/wire"
)

type RouteService interface {
	GetRouteInfo(urlParseDto modelDto.URLParseDTO) (modelDto.RewitePathDTO, error)
}

type routeService struct {
	routeCache component.RouteCache
}

func NewPolicyService(routeCache component.RouteCache) *routeService {
	return &routeService{routeCache: routeCache}
}

func (p *routeService) GetRouteInfo(urlParseDto modelDto.URLParseDTO) (modelDto.RewitePathDTO, error) {
	upstreamService, ok := p.routeCache.Get(urlParseDto.Service)
	if !ok {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching service found for %s", urlParseDto.Service)
	}

	// 서브도메인이 있는 경우를 찾는다.
	domain, emptyDomain := upstreamService.LookupResourceDomain(urlParseDto.Domain)
	if domain == nil {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching domain found for %s", urlParseDto.Domain)
	}

	// URI 경로를 찾는다.
	lookupPath := urlParseDto.GetPath(emptyDomain)
	pathStream := domain.LookupPath(lookupPath)
	if pathStream == nil {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching path found for %s", lookupPath)
	}

	return modelDto.NewRewitePathDTO(pathStream), nil
}

var RouteServiceSet = wire.NewSet(
	NewPolicyService,
	wire.Bind(new(RouteService), new(*routeService)),
)
