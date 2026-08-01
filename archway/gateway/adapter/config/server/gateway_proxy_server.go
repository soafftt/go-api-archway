package server

import (
	"context"
	"core/utils"
	"gateway/adapter/in"
	"gateway/adapter/in/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/wire"
)

var logger = utils.GetLogger()

type GatewayProxyServer struct {
	httpServer *http.Server
}

func NewGatewayProxyServer(
	proxy *in.GatewayProxy,
	containers *middleware.Container,
) *GatewayProxyServer {

	httpServer := http.Server{
		Addr:    ":80",
		Handler: middleware.Chain(proxy.HttpProxy, containers.Middlewares...),
	}

	httpServer.SetKeepAlivesEnabled(true)
	return &GatewayProxyServer{
		httpServer: &httpServer,
	}
}

func (s *GatewayProxyServer) Start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("GatewayProxyServer panic")
			}
		}()

		logger.Info("✅Starting gateway proxy server")
		if err := s.httpServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	defer func() {
		s.registerGracefulShutdown()
	}()
}

func (s *GatewayProxyServer) registerGracefulShutdown() {
	// channel 을 이용하여 stop 시그널을 대기 한다.

	logger.Info("Registering graceful shutdown")

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)
	<-stopSignal

	// 10초간 shutdown 을 대기 한다.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Close(); err != nil {
		// 에러.
		logger.ErrorW("❌http server close error", err)
	}

	select {
	case <-timeoutCtx.Done():
		{
			// timeout 에러
			logger.Error("❌http server close timeout")
		}

	default:
		{
			// 성공
			logger.Info("✅graceful shutdown complete")
		}

	}
}

var ProxyServerProvider = wire.NewSet(NewGatewayProxyServer)
