package di

import (
	"gateway/application/port/in"
	"gateway/application/service"

	"github.com/google/wire"
)

type ApplicationServiceDI struct {
	UpStreamLookUpUseCase in.UpstreamLookupUseCase
}

var ApplicationServiceProvider = wire.NewSet(
	service.NewUpstreamLookupService,
	wire.Struct(new(ApplicationServiceDI), "*"),
)
