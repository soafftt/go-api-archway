//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	"gateway/component"
	"gateway/config"
	"gateway/server"
	middlewareDi "gateway/server/middleware/di"
	serverResponse "gateway/server/response"
	"gateway/service"
	"net/http"
	"net/http/httputil"

	"github.com/google/wire"
)

type GatewayApp struct {
	HttpClient          *http.Client
	Config              *config.AppConfig
	LookupService       service.UpstreamLookupService
	JsonErrorResponse   *serverResponse.JsonErrorResponse
	ReverseServer       *server.ReverseProxyServer
	ReverseProxy        *httputil.ReverseProxy
	MiddlewareContainer *middlewareDi.MiddlewareContainers
}

func InitializeNewApp() (*GatewayApp, error) {
	wire.Build(
		config.AppConfigSet,
		component.HttpClientSet,
		service.UpstreamLookupServiceSet,
		serverResponse.JsonErrorResponseSet,
		middlewareDi.MiddlewareContainerSet,
		server.ReverseProxySet,
		server.ReverseProxyServerSet,
		wire.Struct(new(GatewayApp), "*"),
	)

	return nil, nil
}
