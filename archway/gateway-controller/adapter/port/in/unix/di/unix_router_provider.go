package di

import (
	"gateway/controller/adapter/port/in/unix/handler"

	"github.com/google/wire"
)

type UnixRouterProvider struct {
	UpStreamRouter handler.UpstreamRouter
}

var UnixRouterProviderSet = wire.NewSet(
	handler.NewUpStreamHandler,
	wire.Struct(new(UnixRouterProvider), "*"),
)
