package http

import (
	"core/utils"
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	"sync"
)

var logger = utils.GetLogger()

type HttpMetricOutPort out.ControlPlaneMetricPort

var bodySyncPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 0, 1024*10)
		return &buffer
	},
}

type httpMetric struct {
	httpClient client.HttpClient
}

func NewHttpMetricLookup(
	httpClient client.HttpClient,
) HttpMetricOutPort {
	return &httpMetric{
		httpClient: httpClient,
	}
}

func (u *httpMetric) GetMetric() string {
	resp, err := u.httpClient.GetClient().Get("http://unix/_metrics")
	if err != nil {
		logger.ErrorW("Failed to get metric from control plane", err)
		return ""
	}

	bodyBuffer := bodySyncPool.Get().(*[]byte)
	bodyBytes := *bodyBuffer

	defer func() {
		clear(bodyBytes)
		*bodyBuffer = bodyBytes[:0]
		bodySyncPool.Put(bodyBuffer)

		_ = resp.Body.Close()

	}()

	_, err = resp.Body.Read(bodyBytes)
	if err != nil {
		logger.ErrorW("Failed to get metric from control plane", err)
		return ""
	}

	return utils.ToStringFromBytes(bodyBytes)
}
