package component

import (
	"gateway/config"
	"net"
	"net/http"
	"time"

	"github.com/google/wire"
)

/*
Upstream 의 정보 확인을 위하여, Gateway-Controoller 과 통신하는 HttpClient 를 생성합니다.
*/
func NewHttpClient(appConfig *config.AppConfig) *http.Client {
	httpClientConfig := appConfig.HttpClient

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial(appConfig.Server.Netowrk, appConfig.Server.UnixSocketPath)
		},
		MaxIdleConns:        httpClientConfig.MaxIdleConns,
		MaxIdleConnsPerHost: httpClientConfig.MaxIdleConnsPerHost,
		DisableCompression:  true,
		IdleConnTimeout:     time.Duration(httpClientConfig.IdleConnTimeoutSeconds) * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(httpClientConfig.TimeoutMilliSeconds) * time.Millisecond,
	}
}

var HttpClientSet = wire.NewSet(NewHttpClient)
