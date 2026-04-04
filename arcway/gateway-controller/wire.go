//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	component "gateway/controller/component"
	config "gateway/controller/config"
	infra "gateway/controller/infra"
	server "gateway/controller/server"
	service "gateway/controller/service"

	"github.com/google/wire"
)

type GatewayControllerApp struct {
	Config      *config.AppConfig
	Server      server.GatewayControllerServer
	Service     service.RouteService
	Component   component.ComponentSet
	GlideValkey *infra.ValkeyWrap
}

func InitializeApp() (*GatewayControllerApp, error) {
	wire.Build(
		config.AppConfigSet,
		infra.GlideValkeySet,
		component.RouteComponentSet,
		service.RouteServiceSet,
		server.ServerConfigSet,
		wire.Struct(new(GatewayControllerApp), "*"),
	)
	return nil, nil
}
