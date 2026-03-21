package middleware

import "net/http"

type MiddlewareChain func(next http.Handler) http.Handler

func Chain(h http.Handler, mws ...MiddlewareChain) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
