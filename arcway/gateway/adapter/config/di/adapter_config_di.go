package di

import (
	"gateway/adapter/config"
	"gateway/adapter/config/client"

	"github.com/google/wire"
)

type AdapterConfig struct {
	AppConfig               *config.AppConfig
	GatewayControllerClient client.GatewayControllerClient
}

var AdapterConfigProviderSet = wire.NewSet(
	config.NewAppConfig,
	client.NewGatewayControllerClient,
	wire.Struct(new(AdapterConfig), "*"),
)
