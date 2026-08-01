package middleware

import (
	"core/utils"
	"gateway/adapter/config/appconfig"
	"gateway/adapter/out/controlplane"
	"net/http"
	"strings"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsMiddleware Middleware

type metricsMiddleware struct {
	transfer          string
	grpcMetricOutPort controlplane.GrpcMetricOutPort
	unixMetricOutPort controlplane.UnixMetricOutPort
}

func NewMetricsMiddleware(
	config appconfig.ClientNetworkConfig,
	grpcMetricOutPort controlplane.GrpcMetricOutPort,
	unixMetricOutPort controlplane.UnixMetricOutPort,
) MetricsMiddleware {
	return &metricsMiddleware{
		transfer:          config.GetClientNetworkProperties().Transfer,
		grpcMetricOutPort: grpcMetricOutPort,
		unixMetricOutPort: unixMetricOutPort,
	}
}

var emptyBodyBytes = utils.ToBytesFromString("")

func (m metricsMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URI 자체가 프로메테우스면, 그냥. 넘어감.
		// config 로 URI 를 빼며 될듯.
		switch r.RequestURI {
		// metrics 이면, prometheus 동작.
		case "/_metrics":
			promhttp.Handler().ServeHTTP(w, r)
			break
		case "/_metrics/control-plane":
			w.Header().Set("Content-Type", "text/plain")

			var metricBytes []byte
			var metric string

			if m.transfer == "grpc" {
				metric = m.grpcMetricOutPort.GetMetric()
			} else {
				// http 호출 해야 함.
				metric = m.unixMetricOutPort.GetMetric()
			}

			metric = strings.TrimSpace(metric)
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
			next.ServeHTTP(w, r)
		}
	})
}
