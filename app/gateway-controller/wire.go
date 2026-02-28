//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import (
	config "gateway/controller/config"
	infra "gateway/controller/infra"
	server "gateway/controller/server"

	"github.com/google/wire"
)

type GatewayControllerApp struct {
	Config      *config.AppConfig
	Server      server.GatewayControllerServer
	GlideValkey *infra.GlideValkey
}

func InitializeApp() (*GatewayControllerApp, error) {
	wire.Build(
		config.AppConfigSet,
		infra.GlideValkeySet,
		server.ServerConfigSet,
		wire.Struct(new(GatewayControllerApp), "*"),
	)
	return nil, nil
}
