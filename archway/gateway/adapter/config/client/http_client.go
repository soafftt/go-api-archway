package client

import (
	"context"
	"gateway/adapter/config/appconfig"
	"net"
	"net/http"
	"time"
)

type HttpClientConfig interface {
	GetHttpClientProperties() appconfig.HttpClientProperties
}

type HttpClient interface {
	GetClient() *http.Client
}

type httpClient struct {
	client *http.Client
}

func NewHttpClient(
	networkConfig appconfig.ClientNetworkConfig,
	httpConfig HttpClientConfig,
) HttpClient {
	networkProperties := networkConfig.GetClientNetworkProperties()
	httpClientProperties := httpConfig.GetHttpClientProperties()

	dialer := &net.Dialer{
		Timeout:   time.Duration(httpClientProperties.TimeoutMilliSeconds) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			MaxIdleConns:        500,
			IdleConnTimeout:     90 * time.Second,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, networkProperties.Network, networkProperties.UnixSocketPath)
			},
		},
		Timeout: time.Duration(httpClientProperties.TimeoutMilliSeconds) * time.Millisecond,
	}

	return &httpClient{
		client: &client,
	}
}

func (gcc *httpClient) GetClient() *http.Client {
	return gcc.client
}
