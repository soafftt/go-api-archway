package server

import (
	"gateway/controller/adapter/config/app_config"
	"strings"
)

type ListenerServer interface {
	Start()
}

type GrpcServer ListenerServer
type UnixServer ListenerServer

func NewListenerServer(appConfig *app_config.AppConfig, unixServer UnixServer, rpcServer GrpcServer) ListenerServer {
	if strings.EqualFold(appConfig.Server.Transport, "grpc") {
		return rpcServer
	}

	return unixServer
}
