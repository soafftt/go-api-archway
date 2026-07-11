package client

import (
	"context"
	"gateway/adapter/config"
	"net"
	"net/http"
	"time"
)

type GatewayControllerClient interface {
	GetClient() *http.Client
}

type gatewayControllerClient struct {
	client *http.Client
}

func NewGatewayControllerClient(config *config.AppConfig) GatewayControllerClient {
	dialer := &net.Dialer{
		Timeout:   time.Duration(config.HttpClient.TimeoutMilliSeconds) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, config.GatewayController.Network, config.GatewayController.UNIX_SOCKET_PATH)
			},
		},
		Timeout: time.Duration(config.HttpClient.TimeoutMilliSeconds) * time.Millisecond,
	}

	return &gatewayControllerClient{
		client: &client,
	}
}

func (gcc *gatewayControllerClient) GetClient() *http.Client {
	return gcc.client
}
