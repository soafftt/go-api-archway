package middleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO:  이런건 core 에 만들어 둬야것다.
func Chain(h http.Handler) http.Handler {
	h = metricHandler(h)
	return h
}

func metricHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.RequestURI {
		case "/_metrics":
			promhttp.Handler().ServeHTTP(writer, request)
			break
		default:
			next.ServeHTTP(writer, request)
			break
		}
	})
}
