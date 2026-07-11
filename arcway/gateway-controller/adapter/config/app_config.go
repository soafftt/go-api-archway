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
		MasterHost   string   `env:"VALKEY_MASTER_HOST, require" envDefault:"127.0.0.1:6379"`
		ReplicaHosts []string `env:"VALKEY_REPLICA_HOSTS"`
		ReadFrom     string   `env:"VALKEY_READFROM" envDefault:"master"`
	}
}

func NewAppConfig() *AppConfig {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("env_load fail  %w", err))
	}

	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(fmt.Errorf("env parse fail %w", err))
	}

	if strings.TrimSpace(cfg.Valkey.MasterHost) == "" {
		panic(fmt.Errorf("VALKEY_MASTER_HOSTS contains an empty host"))
	}

	return cfg
}

var AppConfigSet = wire.NewSet(NewAppConfig)
