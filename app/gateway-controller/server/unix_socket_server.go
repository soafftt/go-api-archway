package server

import (
	"context"
	"gateway/controller/config"
	"gateway/controller/router"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/wire"
)

type GatewayControllerServer interface {
	StartUnixSocketServer()
}

type gatewayControllerServer struct {
	config *config.AppConfig
}

func NewGatewayServer(appConfig *config.AppConfig) *gatewayControllerServer {
	return &gatewayControllerServer{config: appConfig}
}

func (g *gatewayControllerServer) StartUnixSocketServer() {
	serverConfig := g.config.Server

	os.Remove(serverConfig.UnixSocketPath)

	listener, err := net.Listen("unix", serverConfig.UnixSocketPath)
	if err != nil {
		panic(err)
	}

	serve := http.Server{
		ReadTimeout:  time.Duration(serverConfig.ReadTimeoutMillisecond) * time.Millisecond,
		WriteTimeout: time.Duration(serverConfig.WriteTimeoutMillisecond) * time.Millisecond,
		IdleTimeout:  time.Duration(serverConfig.IdleTimeoutMillisecond) * time.Millisecond,
		Handler:      router.NewControllerRouter().Mux,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Server Recover", r)
			}
		}()

		log.Printf("UnixSoeketStart: %s", serverConfig.UnixSocketPath)

		if err := serve.Serve(listener); err != nil {
			log.Println("Server Error", err)
		}
	}()

	notifyContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-notifyContext.Done()

	gracefulContext, cancel := context.WithTimeout(notifyContext, 10*time.Second)
	defer cancel()

	log.Println("Shutting down server gracefully...")

	if err := serve.Shutdown(gracefulContext); err != nil {
		log.Printf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Shutting down")
}

var ServerConfigSet = wire.NewSet(
	NewGatewayServer,
	wire.Bind(
		new(GatewayControllerServer), new(*gatewayControllerServer),
	),
)
