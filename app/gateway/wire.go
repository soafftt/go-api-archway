package main

import (
	"gateway/component"
	"gateway/config"
	"net/http"

	"github.com/google/wire"
)

type GatewayApp struct {
	HttpClient *http.Client
	Config     *config.AppConfig
}

func InitializeNewApp() (*GatewayApp, error) {
	wire.Build(
		config.AppConfigSet,
		component.HttpClientSet,
		wire.Struct(new(GatewayApp), "*"),
	)

	return nil, nil
}
