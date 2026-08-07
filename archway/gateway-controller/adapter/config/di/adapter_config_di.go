package di

import (
	appConfig "gateway/controller/adapter/config/app_config"
	"gateway/controller/adapter/config/app_config/server"

	"github.com/google/wire"
)

type AdapterConfig struct {
	AppConfig *appConfig.AppConfig
}

var AdapterConfigProviderSet = wire.NewSet(
	appConfig.NewAppConfig,
	wire.Bind(new(server.HttpServerConfig), new(*appConfig.AppConfig)),
	wire.Bind(new(server.NetworkConfig), new(*appConfig.AppConfig)),
	wire.Bind(new(server.GrpcServerConfig), new(*appConfig.AppConfig)),
	wire.Struct(new(AdapterConfig), "*"),
)
