package server

import (
	"context"
	"gateway/server/middleware"
	"gateway/server/middleware/di"
	"log"
	"net/http"
	"net/http/httputil"
	"os/signal"
	"syscall"

	"github.com/google/wire"
)

type ReverseProxyServer struct {
	proxy                *httputil.ReverseProxy
	middlewareContainers *di.MiddlewareContainers
}

func NewReserveProxyServer(
	gatewayReverseProxy *httputil.ReverseProxy,
	middlewareContainers *di.MiddlewareContainers,
) *ReverseProxyServer {
	return &ReverseProxyServer{
		proxy:                gatewayReverseProxy,
		middlewareContainers: middlewareContainers,
	}
}

func (s *ReverseProxyServer) Start() {
	httpServe := http.Server{
		Addr:    ":80",
		Handler: middleware.Chain(s.proxy, s.middlewareContainers.Middlewares...),
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic occurred: %v", r)
			}
		}()

		log.Printf("Start Server: %s", httpServe.Addr)

		if err := httpServe.ListenAndServe(); err != nil {
			log.Fatalf("서버 실행 실패: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("서버 종료")
}

var ReverseProxyServerSet = wire.NewSet(
	NewReserveProxyServer,
)
