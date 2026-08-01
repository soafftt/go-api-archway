package di

import (
	"gateway/adapter/config/appconfig"
	"gateway/adapter/config/client"

	"github.com/google/wire"
)

type AdapterConfig struct {
	AppConfig  *appconfig.Config
	HttpClient client.HttpClient
	GrpcClient client.GrpcClient
}

var AdapterConfigProviderSet = wire.NewSet(
	appconfig.NewConfig,
	client.NewHttpClient,
	client.NewGrpcClient,
	wire.Bind(new(appconfig.ClientNetworkConfig), new(*appconfig.Config)),
	wire.Bind(new(client.GrpcClientConfig), new(*appconfig.Config)),
	wire.Bind(new(client.HttpClientConfig), new(*appconfig.Config)),
	wire.Struct(new(AdapterConfig), "*"),
)
