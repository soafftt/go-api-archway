package controlplane

import (
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	"sync"
)

// TODO
// grpc metic 도 out.ControlPlaneMetricPort 의 interface 를 추가 해야 함.
// 하나를 가지고 나누어 써야 하고. Golang 의 DI 특성상 같은 interface 를 di graph 에 사용할 수 없기 때문.
type UnixMetricOutPort out.ControlPlaneMetricPort

var bodySyncPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 0, 1024*10)
		return &buffer
	},
}

type unixMetric struct {
	httpClient client.HttpClient
}

func NewUnixMetricLookup(
	httpClient client.HttpClient,
) UnixMetricOutPort {
	return &unixMetric{
		httpClient: httpClient,
	}
}

func (u *unixMetric) GetMetric() string {
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

	return string(bodyBytes)
}
