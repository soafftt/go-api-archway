package grpc_server

import (
	"context"
	"errors"
	"gateway/controller/adapter/config/app_config"
	"gateway/controller/adapter/config/server"
	"gateway/controller/adapter/config/server/grpc_server/metrics"
	adapterPortInGrpcDi "gateway/controller/adapter/port/in/grpc/di"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type grpcServer struct {
	GrpcServiceProvider *adapterPortInGrpcDi.GrpcServiceProvider
	AppConfig           *app_config.AppConfig
	GrpcMetrics         metrics.GrpcServerMetrics
}

func NewGrpcServer(
	appConfig *app_config.AppConfig,
	grpcServiceProvider *adapterPortInGrpcDi.GrpcServiceProvider,
	grpcMetrics metrics.GrpcServerMetrics,
) server.GrpcServer {
	return &grpcServer{
		AppConfig:           appConfig,
		GrpcServiceProvider: grpcServiceProvider,
		GrpcMetrics:         grpcMetrics,
	}
}

func (g *grpcServer) newServer() *grpc.Server {
	grpcConfig := g.AppConfig.Server.Grpc

	numStreamWorkers := grpcConfig.NumStreamWorkers
	if numStreamWorkers == 0 {
		numStreamWorkers = uint32(runtime.GOMAXPROCS(0))
	}

	metrics := g.GrpcMetrics.GetServerMetrics()
	return grpc.NewServer(
		// 최대 ReceiveMessageSize (Bytes)
		grpc.MaxRecvMsgSize(grpcConfig.MaxRecvMsgBytes),
		// 최대 SendMessageSize (Bytes)
		grpc.MaxSendMsgSize(grpcConfig.MaxSendMsgBytes),
		// ReadBufferSize (byte)
		grpc.ReadBufferSize(grpcConfig.ReadBufferBytes),
		// WriteBufferSize (byte)
		grpc.WriteBufferSize(grpcConfig.WriteBufferBytes),
		// ConnectionTime 지정.
		grpc.ConnectionTimeout(time.Duration(grpcConfig.ConnectionTimeoutMillisecond)*time.Millisecond),
		// 회대 연결 동시성 Stream - 1024가 기본이나 2048까지 늘림.
		grpc.MaxConcurrentStreams(grpcConfig.MaxConcurrentStreams),
		// stream 처리를 위한 goworkers - CPU 개수만큼 처리.
		//
		grpc.NumStreamWorkers(numStreamWorkers),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Duration(grpcConfig.KeepaliveMaxConnectionIdleMs) * time.Millisecond,
			MaxConnectionAge:  time.Duration(grpcConfig.KeepaliveMaxConnectionAgeMs) * time.Millisecond,
			Time:              time.Duration(grpcConfig.KeepaliveTimeMs) * time.Millisecond,
			Timeout:           time.Duration(grpcConfig.KeepaliveTimeoutMs) * time.Millisecond,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Duration(grpcConfig.KeepaliveEnforcementMinTimeMs) * time.Millisecond,
			PermitWithoutStream: grpcConfig.PermitWithoutStream,
		}),
		grpc.StreamInterceptor(metrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
}

func (g *grpcServer) registerServices(server grpc.ServiceRegistrar) {
	for i := range g.GrpcServiceProvider.Registrars {
		g.GrpcServiceProvider.Registrars[i].Register(server)
	}
}

func (g *grpcServer) Start() {
	serverConfig := g.AppConfig.Server
	listener, err := net.Listen(serverConfig.Grpc.Network, serverConfig.UnixSocketPath)
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

	shutdownWaitTime := time.Duration(g.AppConfig.Server.Grpc.GracefulStopTimeoutMillisecond) * time.Millisecond
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
