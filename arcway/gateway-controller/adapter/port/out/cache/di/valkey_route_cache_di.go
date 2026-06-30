package di

import (
	adapterOut "gateway/controller/adapter/port/out/cache"
	"gateway/controller/application/port/out/cache"

	"github.com/google/wire"
)

type RouterCacheDi struct {
	RouteCache cache.RouteCache
}

var CacheProviderSet = wire.NewSet(
	adapterOut.NewRouteValkeyCache,
	wire.Bind(new(cache.RouteCache), new(*adapterOut.RouteValkeyCache)),
	wire.Struct(new(RouterCacheDi), "*"),
)
