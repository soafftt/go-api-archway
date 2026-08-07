package unixsocket_server

import (
	"context"
	serverProperties "gateway/controller/adapter/config/app_config/server"
	"gateway/controller/adapter/config/server"
	"gateway/controller/adapter/config/server/unixsocket_server/middleware"
	adapterPortInUnixDi "gateway/controller/adapter/port/in/unix/di"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/wire"
)

type unixServer struct {
	UnixRouterProvider   *adapterPortInUnixDi.UnixRouterProvider
	httpServerProperties serverProperties.HttpServerProperties
	networkProperties    serverProperties.NetworkProperties
}

// Router 를 등록해야 함.
// 즉, Router 를 inject 하여 해야 함.
func NewUnixSocketServer(
	UnixRouterProvider *adapterPortInUnixDi.UnixRouterProvider,
	httpServerConfig serverProperties.HttpServerConfig,
	networkProperties serverProperties.NetworkConfig,
) server.UnixServer {
	return &unixServer{
		UnixRouterProvider:   UnixRouterProvider,
		httpServerProperties: httpServerConfig.GetHttpServerProperties(),
		networkProperties:    networkProperties.GetNetworkProperties(),
	}
}

func (u *unixServer) newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	routes := u.UnixRouterProvider.UpStreamRouter.Routes()
	for i := range routes {
		route := routes[i]
		mux.HandleFunc(route.Method+" "+route.Path, route.Handler)
	}

	return mux
}

func (u *unixServer) newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout:  u.httpServerProperties.ReadTimeoutMillisecond,
		WriteTimeout: u.httpServerProperties.WriteTimeoutMillisecond,
		IdleTimeout:  u.httpServerProperties.IdleTimeoutMillisecond,
		Handler:      middleware.Chain(handler),
	}
}

func (u *unixServer) Start() {
	// unix socket path 를 제거.
	_ = os.Remove(u.networkProperties.UnixSocketPath)

	mux := u.newServeMux()
	serve := u.newHTTPServer(mux)

	listener, err := net.Listen("unix", u.networkProperties.UnixSocketPath)
	if err != nil {
		panic(err)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Server Error", err)
			}
		}()

		log.Printf("Server Started: socket path: %s", u.networkProperties.UnixSocketPath)
		if err := serve.Serve(listener); err != nil {
			log.Println("Server Error", err)
		}
	}()

	// graceful shutdown 처리. ( syscall.SIGINT, syscall.SIGTERM 종료 시그널 수신)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	<-osSignal // 종료 시그널 대기.

	terminateContext, cancel := context.WithTimeout(context.Background(), u.httpServerProperties.GracefulShutdownTimeoutMs)
	defer cancel()

	err = serve.Shutdown(terminateContext)
	switch err {
	case nil:
		log.Println("Graceful Shutdown Complete")
	case context.Canceled:
		log.Printf("Graceful Shutdown Timeout timeout: 10s: %v", err)
	default:
		log.Printf("Graceful Shutdown Fail : %v", err)
	}
}

var UnixServerProvider = wire.NewSet(
	NewUnixSocketServer,
	wire.Struct(new(unixServer), "*"),
)
