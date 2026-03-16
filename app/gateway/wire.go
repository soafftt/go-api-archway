package main

import (
	"gateway/component"
	"gateway/config"
	"gateway/service"
	"net/http"

	"github.com/google/wire"
)

type GatewayApp struct {
	HttpClient    *http.Client
	Config        *config.AppConfig
	LookupService service.UpstreamLookupService
}

func InitializeNewApp() (*GatewayApp, error) {
	wire.Build(
		config.AppConfigSet,
		component.HttpClientSet,
		service.UpstreamLookupServiceSet,
		wire.Struct(new(GatewayApp), "*"),
	)

	return nil, nil
}
