package di

import (
	"gateway/adapter/config"
	"gateway/adapter/config/client"

	"github.com/google/wire"
)

type AdapterConfig struct {
	AppConfig  *config.AppConfig
	HttpClient client.HttpClient
	GrpcClient client.GrpcClient
}

var AdapterConfigProviderSet = wire.NewSet(
	config.NewAppConfig,
	client.NewHttpClient,
	client.NewGrpcClient,
	wire.Struct(new(AdapterConfig), "*"),
)
