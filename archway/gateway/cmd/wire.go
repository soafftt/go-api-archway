//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire
package main

import (
	adpaterConfigDi "gateway/adapter/config/di"
	"gateway/adapter/config/server"
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
	proxyServer        *server.GatewayProxyServer
}

func InitializeApp() (*GatewayProxyApp, error) {
	wire.Build(
		adpaterConfigDi.AdapterConfigProviderSet,
		adapterOutDI.AdapterOutDiProviderSet,
		applicationServiceDi.ApplicationServiceProvider,
		adapterInDI.AdapterInProviderSet,
		server.ProxyServerProvider,
		wire.Struct(new(GatewayProxyApp), "*"),
	)

	return nil, nil
}
