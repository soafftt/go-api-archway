//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	adapderConfigAppConfig "gateway/controller/adapter/config/app_config"
	adapterConfigListener "gateway/controller/adapter/config/server"
	"gateway/controller/adapter/config/server/grpc_server"
	"gateway/controller/adapter/config/server/grpc_server/metrics"
	"gateway/controller/adapter/config/server/unixsocket_server"
	adpaterConfigValkey "gateway/controller/adapter/config/valkey"
	adapterPortInGrpcDi "gateway/controller/adapter/port/in/grpc/di"
	adapterPortInUnixDi "gateway/controller/adapter/port/in/unix/di"
	adapterPortOutDi "gateway/controller/adapter/port/out/di"
	applicationServiceDi "gateway/controller/application/service/di"

	"github.com/google/wire"
)

type GatewayControllerApp struct {
	app                 *adapderConfigAppConfig.AppConfig
	valkeyClient        adpaterConfigValkey.ValkeyClient
	adapterPortOut      *adapterPortOutDi.AdapterPortOutDi
	serviceProvider     *applicationServiceDi.ServiceProvider
	unixRouterProvider  *adapterPortInUnixDi.UnixRouterProvider
	grpcServiceProvider *adapterPortInGrpcDi.GrpcServiceProvider
	unixServer          adapterConfigListener.UnixServer
	grpcServer          adapterConfigListener.GrpcServer
	grpcMetrics         metrics.GrpcServerMetrics
	listenerServer      adapterConfigListener.ListenerServer
}

func InitializeGatewayControllerApp() (*GatewayControllerApp, error) {
	wire.Build(
		adapderConfigAppConfig.AppConfigSet,
		adpaterConfigValkey.ValkeyClientSet,
		adapterPortOutDi.AdapterPortOutDiProviderSet,
		applicationServiceDi.ServiceProviderSet,
		adapterPortInUnixDi.UnixRouterProviderSet,
		adapterPortInGrpcDi.GrpcServiceProviderSet,
		unixsocket_server.UnixServerProvider,
		grpc_server.GrpcServerProvider,
		metrics.GrpcMetricsProvider,
		adapterConfigListener.NewListenerServer,
		wire.Struct(new(GatewayControllerApp), "*"),
	)
	return nil, nil
}
