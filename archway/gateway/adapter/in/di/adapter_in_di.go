package di

import (
	"gateway/adapter/in"
	aid "gateway/adapter/in/middleware"

	"github.com/google/wire"
)

type AdapterInDI struct {
	RequestMiddleware   aid.RequestMiddleware
	MiddlewareContainer *aid.MiddlewareContainer
	GatewayProxy        *in.GatewayProxy
}

var AdapterInProviderSet = wire.NewSet(
	aid.NewRequestMiddleware,
	aid.NewMiddlewareContainer,
	in.NewGatewayProxy,
	wire.Struct(new(AdapterInDI), "*"),
)
