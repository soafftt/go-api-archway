//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire
package main

import (
	"gateway/adapter/config"
	adpaterConfigDi "gateway/adapter/config/di"
	adapterInDI "gateway/adapter/in/di"
	adapterOutDI "gateway/adapter/out/di"
	applicationServiceDi "gateway/application/service/di"

	"github.com/google/wire"
)

type GatewayProxyApp struct {
	adapterConfig      *adpaterConfigDi.AdapterConfig
	adapterOutDI       *adapterOutDI.AdapterOutDI
	adapterInDI        *adapterInDI.AdapterInDI
	applicationService *applicationServiceDi.ApplicationServiceDI
	proxyServer        *config.GatewayProxyServer
}

func InitializeApp() (*GatewayProxyApp, error) {
	wire.Build(
		adpaterConfigDi.AdapterConfigProviderSet,
		adapterOutDI.AdapterOutDiProviderSet,
		applicationServiceDi.ApplicationServiceProvider,
		adapterInDI.AdapterInProviderSet,
		config.ProxyServerProvider,
		wire.Struct(new(GatewayProxyApp), "*"),
	)

	return nil, nil
}
