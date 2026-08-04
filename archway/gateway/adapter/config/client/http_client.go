package client

import (
	"context"
	"gateway/adapter/config/appconfig"
	"net"
	"net/http"
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
		Timeout:   httpClientProperties.TimeoutMilliSeconds,
		KeepAlive: httpClientProperties.KeepAliveSeconds,
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: httpClientProperties.MaxIdleConnsPerHost,
			MaxConnsPerHost:     httpClientProperties.MaxConnsPerHost,
			MaxIdleConns:        httpClientProperties.MaxIdleConns,
			IdleConnTimeout:     httpClientProperties.IdleConnTimeoutSeconds,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, networkProperties.Network, networkProperties.UnixSocketPath)
			},
		},
		Timeout: httpClientProperties.TimeoutMilliSeconds,
	}

	return &httpClient{
		client: &client,
	}
}

func (gcc *httpClient) GetClient() *http.Client {
	return gcc.client
}
