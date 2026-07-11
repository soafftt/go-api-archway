package di

import (
	"gateway/controller/adapter/port/out/cache/di"

	"github.com/google/wire"
)

type AdapterPortOutDi struct {
	RouterCache di.RouterCacheDi
}

var AdapterPortOutDiProviderSet = wire.NewSet(
	di.CacheProviderSet,
	wire.Struct(new(AdapterPortOutDi), "*"),
)
