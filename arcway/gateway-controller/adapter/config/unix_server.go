package config

import (
	"context"
	adapterPortInUnixDi "gateway/controller/adapter/port/in/unix/di"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/wire"
)

type UnixServer struct {
	UnixRouterProvider *adapterPortInUnixDi.UnixRouterProvider
	AppConfig          *AppConfig
}

// Router 를 등록해야 함.
// 즉, Router 를 inject 하여 해야 함.

func (u *UnixServer) newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	routes := u.UnixRouterProvider.UpStreamRouter.Routes()
	for i := range routes {
		route := routes[i]
		mux.HandleFunc(route.Method+" "+route.Path, route.Handler)
	}

	return mux
}

func (u *UnixServer) newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		ReadTimeout:  time.Duration(u.AppConfig.Server.ReadTimeoutMillisecond) * time.Millisecond,
		WriteTimeout: time.Duration(u.AppConfig.Server.WriteTimeoutMillisecond) * time.Millisecond,
		IdleTimeout:  time.Duration(u.AppConfig.Server.IdleTimeoutMillisecond) * time.Millisecond,
		Handler:      handler,
	}
}

func (u *UnixServer) Start() {
	// unix socket path 를 제거.
	_ = os.Remove(u.AppConfig.Server.UnixSocketPath)

	mux := u.newServeMux()
	serve := u.newHTTPServer(mux)

	listener, err := net.Listen("unix", u.AppConfig.Server.UnixSocketPath)
	if err != nil {
		panic(err)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Server Error", err)
			}
		}()

		log.Printf("Server Started: socket path: %s", u.AppConfig.Server.UnixSocketPath)
		if err := serve.Serve(listener); err != nil {
			log.Println("Server Error", err)
		}
	}()

	// graceful shutdown 처리. ( syscall.SIGINT, syscall.SIGTERM 종료 시그널 수신)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	<-osSignal // 종료 시그널 대기.

	terminateContext, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
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

var UnixServerProvider = wire.Struct(new(UnixServer), "*")
