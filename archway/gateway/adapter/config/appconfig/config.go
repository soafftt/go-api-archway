package appconfig

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type ClientNetworkConfig interface {
	GetClientNetworkProperties() ClientNetworkProperties
}

type Config struct {
	ClientNetworkConfig ClientNetworkProperties
	GrpcClientConfig    GrpcClientProperties
	HttpClientConfig    HttpClientProperties
}

func (c Config) GetHttpClientProperties() HttpClientProperties {
	return c.HttpClientConfig
}

func (c Config) GetClientNetworkProperties() ClientNetworkProperties {
	return c.ClientNetworkConfig
}

func (c Config) GetGrpcClientProperties() GrpcClientProperties {
	return c.GrpcClientConfig
}

type ClientNetworkProperties struct {
	Transfer          string `env:"GATEWAY_CONTROLLER_TRANSFER,required" envDefault:"grpc"`
	UnixSocketBaseURI string `env:"GATEWAY_CONTROLLER_UPSTREAM_LOOKUP_UNIX_SOCKET_BASE_URI" envDefault:"http://unix/v1/upstream?path="`
	Network           string `env:"GATEWAY_CONTROLLER_SERVER_NETWORK" envDefault:"unix"`
	UnixSocketPath    string `env:"GATEWAY_CONTROLLER_UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
}

type GrpcClientProperties struct {
	MaxRecvMsgSize      int           `env:"GRPC_CLIENT_MAX_RECV_MSG_SIZE" envDefault:"2147483647"`
	KeepaliveTime       time.Duration `env:"GRPC_CLIENT_KEEPALIVE_TIME" envDefault:"10s"`
	KeepaliveTimeout    time.Duration `env:"GRPC_CLIENT_KEEPALIVE_TIMEOUT" envDefault:"1s"`
	PermitWithoutStream bool          `env:"GRPC_CLIENT_PERMIT_WITHOUT_STREAM" envDefault:"true"`
	MinConnectTimeout   time.Duration `env:"GRPC_CLIENT_MIN_CONNECT_TIMEOUT" envDefault:"1s"`
	BackoffBaseDelay    time.Duration `env:"GRPC_CLIENT_BACKOFF_BASE_DELAY" envDefault:"1s"`
	BackoffMultiplier   float64       `env:"GRPC_CLIENT_BACKOFF_MULTIPLIER" envDefault:"1.2"`
	BackoffJitter       float64       `env:"GRPC_CLIENT_BACKOFF_JITTER" envDefault:"0.2"`
	BackoffMaxDelay     time.Duration `env:"GRPC_CLIENT_BACKOFF_MAX_DELAY" envDefault:"2s"`
	DisableRetry        bool          `env:"GRPC_CLIENT_DISABLE_RERY" envDefault:"false"`
}

type HttpClientProperties struct {
	MaxIdleConns           int `env:"HTTP_CLIENT_MAX_IDLE_CONNS" envDefault:"250"`
	MaxIdleConnsPerHost    int `env:"HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST" envDefault:"500"`
	IdleConnTimeoutSeconds int `env:"HTTP_CLIENT_IDLE_CONN_TIMEOUT" envDefault:"90"`
	TimeoutMilliSeconds    int `env:"HTTP_CLIENT_TIMEOUT_MILLISECONDS" envDefault:"5000"`
}

func NewConfig() *Config {
	_ = godotenv.Load()
	var cfg = &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
