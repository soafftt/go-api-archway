package middleware

import (
	"core/utils"
	"gateway/adapter/config/appconfig"
	"gateway/adapter/out/controlplane/grpc"
	http2 "gateway/adapter/out/controlplane/http"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsMiddleware Middleware

type metricsMiddleware struct {
	transfer          string
	grpcMetricOutPort grpc.GrpcMetricOutPort
	unixMetricOutPort http2.HttpMetricOutPort
}

func NewMetricsMiddleware(
	config appconfig.ClientNetworkConfig,
	grpcMetricOutPort grpc.GrpcMetricOutPort,
	unixMetricOutPort http2.HttpMetricOutPort,
) MetricsMiddleware {
	return &metricsMiddleware{
		transfer:          config.GetClientNetworkProperties().Transfer,
		grpcMetricOutPort: grpcMetricOutPort,
		unixMetricOutPort: unixMetricOutPort,
	}
}

func (m metricsMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URI 자체가 프로메테우스면, 그냥. 넘어감.
		switch r.RequestURI {
		// metrics 이면, prometheus 동작.
		case "/_metrics":
			promhttp.Handler().ServeHTTP(w, r)
			break
		case "/_metrics/control-plane":
			metricBytes := m.getMetric()

			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write(metricBytes)

			break
		case "/favicon.ico":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write(utils.GetEmptyStringBytes())
			break

		// 기본 통작.
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (m metricsMiddleware) getMetric() []byte {
	var metric string

	if m.transfer == "grpc" {
		metric = m.grpcMetricOutPort.GetMetric()
	} else {
		// http 호출 해야 함.
		metric = m.unixMetricOutPort.GetMetric()
	}

	metric = strings.TrimSpace(metric)
	if metric == "" {
		return utils.GetEmptyStringBytes()
	}

	return utils.ToBytesFromString(metric)
}
