package server

import (
	serverConfig "gateway/controller/adapter/config/app_config/server"
	"strings"
)

type ListenerServer interface {
	Start()
}

type GrpcServer ListenerServer
type UnixServer ListenerServer

func NewListenerServer(networkConfig serverConfig.NetworkConfig, unixServer UnixServer, rpcServer GrpcServer) ListenerServer {
	if strings.EqualFold(networkConfig.GetNetworkProperties().Transport, "grpc") {
		return rpcServer
	}

	return unixServer
}
