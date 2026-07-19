package app_config

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/google/wire"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server struct {
		Transport                 string `env:"SERVER_TRANSPORT" envDefault:"grpc"`
		UnixSocketPath            string `env:"UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
		ReadTimeoutMillisecond    int64  `env:"READ_TIMEOUT_MILLISECOND" envDefault:"10"`
		WriteTimeoutMillisecond   int64  `env:"WRITE_TIMEOUT_MILLISECOND" envDefault:"10"`
		IdleTimeoutMillisecond    int64  `env:"IDLE_TIMEOUT_MILLISECOND" envDefault:"120"`
		GracefulShutdownTimeoutMs int64  `env:"GRACEFUL_SHUTDOWN_TIMEOUT_MILLISECOND" envDefault:"10000"`
		Grpc                      struct {
			Network                        string `env:"GRPC_NETWORK" envDefault:"unix"`
			MaxRecvMsgBytes                int    `env:"GRPC_MAX_RECV_MSG_BYTES" envDefault:"4194304"`
			MaxSendMsgBytes                int    `env:"GRPC_MAX_SEND_MSG_BYTES" envDefault:"10485760"`
			ReadBufferBytes                int    `env:"GRPC_READ_BUFFER_BYTES" envDefault:"32768"`
			WriteBufferBytes               int    `env:"GRPC_WRITE_BUFFER_BYTES" envDefault:"32768"`
			ConnectionTimeoutMillisecond   int64  `env:"GRPC_CONNECTION_TIMEOUT_MILLISECOND" envDefault:"5000"`
			MaxConcurrentStreams           uint32 `env:"GRPC_MAX_CONCURRENT_STREAMS" envDefault:"2048"`
			NumStreamWorkers               uint32 `env:"GRPC_NUM_STREAM_WORKERS" envDefault:"0"`
			KeepaliveMaxConnectionIdleMs   int64  `env:"GRPC_KEEPALIVE_MAX_CONNECTION_IDLE_MILLISECOND" envDefault:"900000"`
			KeepaliveMaxConnectionAgeMs    int64  `env:"GRPC_KEEPALIVE_MAX_CONNECTION_AGE_MILLISECOND" envDefault:"1800000"`
			KeepaliveTimeMs                int64  `env:"GRPC_KEEPALIVE_TIME_MILLISECOND" envDefault:"120000"`
			KeepaliveTimeoutMs             int64  `env:"GRPC_KEEPALIVE_TIMEOUT_MILLISECOND" envDefault:"20000"`
			KeepaliveEnforcementMinTimeMs  int64  `env:"GRPC_KEEPALIVE_ENFORCEMENT_MIN_TIME_MILLISECOND" envDefault:"20000"`
			PermitWithoutStream            bool   `env:"GRPC_KEEPALIVE_PERMIT_WITHOUT_STREAM" envDefault:"true"`
			GracefulStopTimeoutMillisecond int64  `env:"GRPC_GRACEFUL_STOP_TIMEOUT_MILLISECOND" envDefault:"10000"`
		}
	}

	Valkey struct {
		MasterHost   string   `env:"VALKEY_MASTER_HOST,required" envDefault:"127.0.0.1:6379"`
		ReplicaHosts []string `env:"VALKEY_REPLICA_HOSTS"`
		ReadFrom     string   `env:"VALKEY_READFROM" envDefault:"master"`
	}
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

var AppConfigSet = wire.NewSet(NewAppConfig)
