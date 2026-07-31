package middleware

import (
	"gateway/adapter/config"
	"gateway/application/port/in"
	"net/http"
	"strings"
	"unsafe"
)

type MetricsMiddleware Middleware

type metricsMiddleware struct {
	transfer          string
	grpcMetricUseCase in.ControlPlaneMetricUseCase
}

func NewMetricsMiddleware(
	config *config.AppConfig,
	controlPlaneMetricUseCase in.ControlPlaneMetricUseCase,
) MetricsMiddleware {
	return &metricsMiddleware{
		transfer:          config.ClientNetworkConfig.Transfer,
		grpcMetricUseCase: controlPlaneMetricUseCase,
	}
}

var emptyBodyBytes = []byte("")

func (m metricsMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URI 자체가 프로메테우스면, 그냥. 넘어감.
		// config 로 URI 를 빼며 될듯.
		switch r.RequestURI {
		// metrics 이면, prometheus 동작.
		case "/_metrics":
			break
		case "/_metrics/control-plane":
			metric := strings.TrimSpace(m.grpcMetricUseCase.GetMetric())
			w.Header().Set("Content-Type", "text/plain")

			var metricBytes []byte
			if metric == "" {
				metricBytes = emptyBodyBytes
			} else {
				metricBytes = unsafe.Slice(unsafe.StringData(metric), len(metric))
			}

			_, _ = w.Write(metricBytes)
			break
		case "/favicon.ico":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write(emptyBodyBytes)
			break

		// 기본 통작.
		default:
			// prometheus 설정을 찾아보자.
			next.ServeHTTP(w, r)
		}
	})
}
