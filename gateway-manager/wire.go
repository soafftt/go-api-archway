//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package main

import "github.com/google/wire"

type GetwayManager struct {
	Config *Config
}

func InitializeApp() (*GetwayManager, error) {
	wire.Build(
		ConfigSet,
		wire.Struct(new(GetwayManager), "*"),
	)
	return nil, nil
}
