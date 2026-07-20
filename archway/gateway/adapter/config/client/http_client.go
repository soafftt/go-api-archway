package client

import (
	"context"
	"gateway/adapter/config"
	"net"
	"net/http"
	"time"
)

type HttpClient interface {
	GetClient() *http.Client
}

type httpClient struct {
	client *http.Client
}

func NewHttpClient(config *config.AppConfig) HttpClient {
	dialer := &net.Dialer{
		Timeout:   time.Duration(config.HttpClient.TimeoutMilliSeconds) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			MaxIdleConns:        500,
			IdleConnTimeout:     90 * time.Second,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, config.ClientNetworkConfig.Network, config.ClientNetworkConfig.UnixSocketPath)
			},
		},
		Timeout: time.Duration(config.HttpClient.TimeoutMilliSeconds) * time.Millisecond,
	}

	return &httpClient{
		client: &client,
	}
}

func (gcc *httpClient) GetClient() *http.Client {
	return gcc.client
}
