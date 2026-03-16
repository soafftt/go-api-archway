package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/google/wire"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server struct {
		Netowrk        string `env:"SERVER_NETWORK" envDefault:"unix"`
		UnixSocketPath string `env:"UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
	}
	HttpClient struct {
		MaxIdleConns           int `env:"HTTP_CLIENT_MAX_IDLE_CONNS" envDefault:"0"`
		MaxIdleConnsPerHost    int `env:"HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST" envDefault:"10000"`
		IdleConnTimeoutSeconds int `env:"HTTP_CLIENT_IDLE_CONN_TIMEOUT" envDefault:"90"`
		TimeoutMilliSeconds    int `env:"HTTP_CLIENT_TIMEOUT_MILLISECONDS" envDefault:"5000"`
	}
	UpstreamLookup struct {
		BaseURL string `env:"UPSTREAM_LOOKUP_BASE_URL" envDefault:"http://localhost/v1/upstream?path="`
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
