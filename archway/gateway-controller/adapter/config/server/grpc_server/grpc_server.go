package grpc_server

import (
	"context"
	"errors"
	serverProperties "gateway/controller/adapter/config/app_config/server"
	"gateway/controller/adapter/config/server"
	"gateway/controller/adapter/config/server/grpc_server/metrics"
	adapterPortInGrpcDi "gateway/controller/adapter/port/in/grpc/di"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type grpcServer struct {
	GrpcServiceProvider  *adapterPortInGrpcDi.GrpcServiceProvider
	networkProperties    serverProperties.NetworkProperties
	grpcServerProperties serverProperties.GrpcServerProperties
	GrpcMetrics          metrics.GrpcServerMetrics
}

func NewGrpcServer(
	networkConfig serverProperties.NetworkConfig,
	grpcConfig serverProperties.GrpcServerConfig,
	grpcServiceProvider *adapterPortInGrpcDi.GrpcServiceProvider,
	grpcMetrics metrics.GrpcServerMetrics,
) server.GrpcServer {
	return &grpcServer{
		networkProperties:    networkConfig.GetNetworkProperties(),
		grpcServerProperties: grpcConfig.GetGrpcServerProperties(),
		GrpcServiceProvider:  grpcServiceProvider,
		GrpcMetrics:          grpcMetrics,
	}
}

func (g *grpcServer) newServer() *grpc.Server {
	numStreamWorkers := g.grpcServerProperties.NumStreamWorkers
	if numStreamWorkers == 0 {
		numStreamWorkers = uint32(runtime.GOMAXPROCS(0))
	}

	serverMetrics := g.GrpcMetrics.GetServerMetrics()
	return grpc.NewServer(
		// 최대 ReceiveMessageSize (Bytes)
		grpc.MaxRecvMsgSize(g.grpcServerProperties.MaxRecvMsgBytes),
		// 최대 SendMessageSize (Bytes)
		grpc.MaxSendMsgSize(g.grpcServerProperties.MaxSendMsgBytes),
		// ReadBufferSize (byte)
		grpc.ReadBufferSize(g.grpcServerProperties.ReadBufferBytes),
		// WriteBufferSize (byte)
		grpc.WriteBufferSize(g.grpcServerProperties.WriteBufferBytes),
		// ConnectionTime 지정.
		grpc.ConnectionTimeout(g.grpcServerProperties.ConnectionTimeoutMillisecond),
		// 회대 연결 동시성 Stream - 1024가 기본이나 2048까지 늘림.
		grpc.MaxConcurrentStreams(g.grpcServerProperties.MaxConcurrentStreams),
		// stream 처리를 위한 goworkers - CPU 개수만큼 처리.
		//
		grpc.NumStreamWorkers(numStreamWorkers),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: g.grpcServerProperties.KeepaliveMaxConnectionIdleMs,
			MaxConnectionAge:  g.grpcServerProperties.KeepaliveMaxConnectionAgeMs,
			Time:              g.grpcServerProperties.KeepaliveTimeMs,
			Timeout:           g.grpcServerProperties.KeepaliveTimeoutMs,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             g.grpcServerProperties.KeepaliveEnforcementMinTimeMs,
			PermitWithoutStream: g.grpcServerProperties.PermitWithoutStream,
		}),
		grpc.StreamInterceptor(serverMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(serverMetrics.UnaryServerInterceptor()),
	)
}

func (g *grpcServer) registerServices(server grpc.ServiceRegistrar) {
	for i := range g.GrpcServiceProvider.Registrars {
		g.GrpcServiceProvider.Registrars[i].Register(server)
	}
}

func (g *grpcServer) Start() {
	listener, err := net.Listen(g.grpcServerProperties.Network, g.networkProperties.UnixSocketPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = listener.Close()
	}()

	rpcServer := g.newServer()
	g.registerServices(rpcServer)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()

		log.Println("grpc server start")
		if err := rpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Println(err)
		}
	}()

	g.stopWait(rpcServer)

}

func (g *grpcServer) stopWait(server *grpc.Server) {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(osSignal)

	<-osSignal

	shutdownWaitTime := g.grpcServerProperties.GracefulStopTimeoutMillisecond
	timeoutContext, cancel := context.WithTimeout(context.Background(), shutdownWaitTime)
	defer cancel()

	shutdownDone := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		log.Println("gRPC graceful shutdown complete")
	case <-timeoutContext.Done():
		server.Stop()
		if errors.Is(timeoutContext.Err(), context.DeadlineExceeded) {
			log.Printf("gRPC graceful shutdown timeout: %s", shutdownWaitTime)
		} else {
			log.Printf("gRPC graceful shutdown canceled: %v", timeoutContext.Err())
		}
	}
}

var GrpcServerProvider = wire.NewSet(
	NewGrpcServer,
	wire.Struct(new(grpcServer), "*"),
)
