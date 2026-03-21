//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	"gateway/component"
	"gateway/config"
	"gateway/server"
	"gateway/service"
	"net/http"

	"github.com/google/wire"
)

type GatewayApp struct {
	HttpClient    *http.Client
	Config        *config.AppConfig
	LookupService service.UpstreamLookupService
	ReverseProxy  *server.ReverseProxyServer
}

func InitializeNewApp() (*GatewayApp, error) {
	wire.Build(
		config.AppConfigSet,
		component.HttpClientSet,
		service.UpstreamLookupServiceSet,
		server.ReverseProxySet,
		wire.Struct(new(GatewayApp), "*"),
	)

	return nil, nil
}
