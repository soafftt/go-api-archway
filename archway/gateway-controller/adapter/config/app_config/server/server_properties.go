package server

import "time"

type NetworkConfig interface {
	GetNetworkProperties() NetworkProperties
}

type NetworkProperties struct {
	Transport      string `env:"SERVER_TRANSPORT" envDefault:"grpc"`
	UnixSocketPath string `env:"UNIX_SOCKET_PATH" envDefault:"/tmp/gateway-controller.sock"`
}

type HttpServerConfig interface {
	GetHttpServerProperties() HttpServerProperties
}

type HttpServerProperties struct {
	ReadTimeoutMillisecond    time.Duration `env:"READ_TIMEOUT_MILLISECOND" envDefault:"10ms"`
	WriteTimeoutMillisecond   time.Duration `env:"WRITE_TIMEOUT_MILLISECOND" envDefault:"10ms"`
	IdleTimeoutMillisecond    time.Duration `env:"IDLE_TIMEOUT_MILLISECOND" envDefault:"120ms"`
	GracefulShutdownTimeoutMs time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT_MILLISECOND" envDefault:"10000ms"`
}

type GrpcServerConfig interface {
	GetGrpcServerProperties() GrpcServerProperties
}

type GrpcServerProperties struct {
	Network                        string        `env:"GRPC_NETWORK" envDefault:"unix"`
	MaxRecvMsgBytes                int           `env:"GRPC_MAX_RECV_MSG_BYTES" envDefault:"4194304"`
	MaxSendMsgBytes                int           `env:"GRPC_MAX_SEND_MSG_BYTES" envDefault:"10485760"`
	ReadBufferBytes                int           `env:"GRPC_READ_BUFFER_BYTES" envDefault:"32768"`
	WriteBufferBytes               int           `env:"GRPC_WRITE_BUFFER_BYTES" envDefault:"32768"`
	ConnectionTimeoutMillisecond   time.Duration `env:"GRPC_CONNECTION_TIMEOUT_MILLISECOND" envDefault:"5000ms"`
	MaxConcurrentStreams           uint32        `env:"GRPC_MAX_CONCURRENT_STREAMS" envDefault:"2048"`
	NumStreamWorkers               uint32        `env:"GRPC_NUM_STREAM_WORKERS" envDefault:"0"`
	KeepaliveMaxConnectionIdleMs   time.Duration `env:"GRPC_KEEPALIVE_MAX_CONNECTION_IDLE_MILLISECOND" envDefault:"900000ms"`
	KeepaliveMaxConnectionAgeMs    time.Duration `env:"GRPC_KEEPALIVE_MAX_CONNECTION_AGE_MILLISECOND" envDefault:"1800000ms"`
	KeepaliveTimeMs                time.Duration `env:"GRPC_KEEPALIVE_TIME_MILLISECOND" envDefault:"120000ms"`
	KeepaliveTimeoutMs             time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT_MILLISECOND" envDefault:"20000ms"`
	KeepaliveEnforcementMinTimeMs  time.Duration `env:"GRPC_KEEPALIVE_ENFORCEMENT_MIN_TIME_MILLISECOND" envDefault:"20000ms"`
	PermitWithoutStream            bool          `env:"GRPC_KEEPALIVE_PERMIT_WITHOUT_STREAM" envDefault:"true"`
	GracefulStopTimeoutMillisecond time.Duration `env:"GRPC_GRACEFUL_STOP_TIMEOUT_MILLISECOND" envDefault:"10000ms"`
}
