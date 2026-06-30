package di

import (
	"gateway/controller/application/port/in"
	"gateway/controller/application/service"

	"github.com/google/wire"
)

type ServiceProvider struct {
	UpstreamLookupCase in.UpstreamLookupUseCase
}

var ServiceProviderSet = wire.NewSet(
	service.NewUpstreamLookupService,
	wire.Struct(new(ServiceProvider), "*"),
)
