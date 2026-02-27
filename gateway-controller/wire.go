//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import "github.com/google/wire"

type GatewayControllerApp struct {
	Config *Config
}

func InitializeApp() (*GatewayControllerApp, error) {
	wire.Build(
		ConfigSet,
		wire.Struct(new(GatewayControllerApp), "*"),
	)
	return nil, nil
}
