package middleware

import (
	"net/http"
)

// 이걸 구현하게 만들고.
type Middleware interface {
	HandleMiddleware(next http.Handler) http.Handler
}

// serve 를 구현하는 곳에서 chain 실행.
// 내가 등록한 순서대로의 hm ...Middleware 을 역순으로 등록해야 순차적으로 실행됨.
func Chain(h http.Handler, hm ...Middleware) http.Handler {
	for i := len(hm) - 1; i >= 0; i-- {
		// 역순으로 미들웨어를 등록하면, 정순으로 실행된다.
		h = hm[i].HandleMiddleware(h)
	}
	return h
}

// MiddlewareCotainer 를 만들고, DI 과정에서 내가 필요로 하는는 것을 순서대로 등록함.
type MiddlewareContainer struct {
	Middlewares []Middleware
}

func NewMiddlewareContainer(requestMiddleware RequestMiddleware) *MiddlewareContainer {
	middlewares := make([]Middleware, 0, 1)
	middlewares = append(middlewares, requestMiddleware)

	return &MiddlewareContainer{
		Middlewares: middlewares,
	}
}
