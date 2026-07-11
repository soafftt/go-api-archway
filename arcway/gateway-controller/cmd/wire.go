//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	adpaterConfig "gateway/controller/adapter/config"
	adapterPortInUnixDi "gateway/controller/adapter/port/in/unix/di"
	adapterPortOutDi "gateway/controller/adapter/port/out/di"
	applicationServiceDi "gateway/controller/application/service/di"

	"github.com/google/wire"
)

type GatewayControllerApp struct {
	app                *adpaterConfig.AppConfig
	valkeyClient       adpaterConfig.ValkeyClient
	adapterPortOut     *adapterPortOutDi.AdapterPortOutDi
	serviceProvider    *applicationServiceDi.ServiceProvider
	unixRouterProvider *adapterPortInUnixDi.UnixRouterProvider
	unixServer         *adpaterConfig.UnixServer
}

func InitializeGatewayControllerApp() (*GatewayControllerApp, error) {
	wire.Build(
		adpaterConfig.AppConfigSet,
		adpaterConfig.ValkeyClientSet,
		adapterPortOutDi.AdapterPortOutDiProviderSet,
		applicationServiceDi.ServiceProviderSet,
		adapterPortInUnixDi.UnixRouterProviderSet,
		adpaterConfig.UnixServerProvider,
		wire.Struct(new(GatewayControllerApp), "*"),
	)
	return nil, nil
}
