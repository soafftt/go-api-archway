package app_config

import (
	"fmt"
	"gateway/controller/adapter/config/app_config/server"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server struct {
		NetworkProperties    server.NetworkProperties
		HttpServerProperties server.HttpServerProperties
		GrpcServerProperties server.GrpcServerProperties
	}
	Valkey struct {
		MasterHost   string   `env:"VALKEY_MASTER_HOST,required" envDefault:"127.0.0.1:6379"`
		ReplicaHosts []string `env:"VALKEY_REPLICA_HOSTS"`
		ReadFrom     string   `env:"VALKEY_READFROM" envDefault:"master"`
	}
}

func (a AppConfig) GetGrpcServerProperties() server.GrpcServerProperties {
	return a.Server.GrpcServerProperties
}

func (a AppConfig) GetNetworkProperties() server.NetworkProperties {
	return a.Server.NetworkProperties
}

func (a AppConfig) GetHttpServerProperties() server.HttpServerProperties {
	return a.Server.HttpServerProperties
}

func NewAppConfig() *AppConfig {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
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
