package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/google/wire"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server struct {
		UnixSocketPath          string `env:"UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
		ReadTimeoutMillisecond  int64  `env:"READ_TIMEOUT_MILLISECOND" envDefault:"10"`
		WriteTimeoutMillisecond int64  `env:"WRITE_TIMEOUT_MILLISECOND" envDefault:"10"`
		IdleTimeoutMillisecond  int64  `env:"IDLE_TIMEOUT_MILLISECOND" envDefault:"120"`
	}

	Valkey struct {
		Hosts    []string `env:"VALKEY_HOSTS,required" envSeparator:","`
		ReadFrom string   `env:"VALKEY_READFROM" envDefault:"master"`
	}
}

func NewAppConfig() *AppConfig {
	godotenv.Load()

	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	for _, host := range cfg.Valkey.Hosts {
		if strings.TrimSpace(host) == "" {
			panic(fmt.Errorf("VALKEY_HOSTS contains an empty host"))
		}
	}

	return cfg
}

var AppConfigSet = wire.NewSet(NewAppConfig)
