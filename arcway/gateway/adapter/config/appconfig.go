package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	GatewayController struct {
		BaseURL          string `env:"GATEWAY_CONTROLLER_UPSTREAM_LOOKUP_BASE_URL" envDefault:"http://unix/v1/upstream?path="`
		Network          string `env:"GATEWAY_CONTROLLER_SERVER_NETWORK" envDefault:"unix"`
		UNIX_SOCKET_PATH string `env:"GATEWAY_CONTROLLER_UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
	}
	HttpClient struct {
		MaxIdleConns           int `env:"HTTP_CLIENT_MAX_IDLE_CONNS" envDefault:"250"`
		MaxIdleConnsPerHost    int `env:"HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST" envDefault:"500"`
		IdleConnTimeoutSeconds int `env:"HTTP_CLIENT_IDLE_CONN_TIMEOUT" envDefault:"90"`
		TimeoutMilliSeconds    int `env:"HTTP_CLIENT_TIMEOUT_MILLISECONDS" envDefault:"5000"`
	}
}

func NewAppConfig() *AppConfig {
	_ = godotenv.Load()

	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
