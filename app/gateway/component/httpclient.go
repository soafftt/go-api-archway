package component

import (
	"net"
	"net/http"
	"time"

	"github.com/google/wire"
)

func NewHttpClient() *http.Client {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", "socketPath")
		},
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 10000,
		DisableCompression:  true,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Millisecond,
	}
}

var HttpClientSet = wire.NewSet(NewHttpClient)
