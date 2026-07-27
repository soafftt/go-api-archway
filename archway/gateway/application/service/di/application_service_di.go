package di

import (
	"gateway/application/port/in"
	"gateway/application/service"

	"github.com/google/wire"
)

type ApplicationServiceDI struct {
	UpStreamLookUpUseCase     in.UpstreamLookupUseCase
	ControlPlaneMetricUseCase in.ControlPlaneMetricUseCase
}

var ApplicationServiceProvider = wire.NewSet(
	service.NewUpstreamLookupService,
	service.NewControlPlaneMetricService,
	wire.Struct(new(ApplicationServiceDI), "*"),
)
