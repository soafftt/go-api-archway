package di

import (
	"gateway/adapter/in"
	aid "gateway/adapter/in/middleware"

	"github.com/google/wire"
)

type AdapterInDI struct {
	RequestMiddleware   aid.RequestMiddleware
	MetricsMiddleware   aid.MetricsMiddleware
	MiddlewareContainer *aid.Container
	GatewayProxy        *in.GatewayProxy
}

var AdapterInProviderSet = wire.NewSet(
	aid.NewRequestMiddleware,
	aid.NewMetricsMiddleware,
	aid.NewMiddlewareContainer,
	in.NewGatewayProxy,
	wire.Struct(new(AdapterInDI), "*"),
)
