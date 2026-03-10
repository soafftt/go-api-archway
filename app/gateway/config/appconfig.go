package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/google/wire"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server struct {
		UnixSocketPath string `env:"UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
	}
}

func NewAppConfig() *AppConfig {
	godotenv.Load()

	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}

var AppConfigSet = wire.NewSet(NewAppConfig)
