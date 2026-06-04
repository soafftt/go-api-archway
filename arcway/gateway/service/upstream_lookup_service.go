package service

import (
	"gateway/model"
	"gateway/port/outbound"

	"github.com/google/wire"
)

type UpstreamLookupService interface {
	Lookup(targetPath string) model.UpstreamLookupResult
}

type upstreamLookupService struct {
	upstreamLookupPort outbound.UpstreamLookupPort
}

func NewUpstreamLookupService(upstreamLookupPort outbound.UpstreamLookupPort) *upstreamLookupService {
	return &upstreamLookupService{
		upstreamLookupPort: upstreamLookupPort,
	}
}

/*
Upstream 에 대한 정보를 조회합니다.
*/
func (s *upstreamLookupService) Lookup(targetPath string) model.UpstreamLookupResult {
	return s.upstreamLookupPort.Lookup(targetPath)
}

var UpstreamLookupServiceSet = wire.NewSet(
	NewUpstreamLookupService,
	wire.Bind(new(UpstreamLookupService), new(*upstreamLookupService)),
)
