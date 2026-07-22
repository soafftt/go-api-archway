package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsMiddleware Middleware

type metricsMiddleware struct{}

func NewMetricsMiddleware() MetricsMiddleware {
	return &metricsMiddleware{}
}

func (m metricsMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URI 자체가 프로메테우스면, 그냥. 넘어감.
		// config 로 URI 를 빼며 될듯.
		switch r.RequestURI {
		// metrics 이면, prometheus 동작.
		case "/_metrics":
			promhttp.Handler().ServeHTTP(w, r)
			break
		case "/favicon.ico":
			{
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte(""))
				break
			}

		// 기본 통작.
		default:
			{
				// prometheus 설정을 찾아보자.
				next.ServeHTTP(w, r)
			}
		}
	})

}
