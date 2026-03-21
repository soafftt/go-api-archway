package server

import "net/http"

// 함수 type 정의
type Middleware interface {
	handleMiddleware(next http.Handler) http.Handler
}

func Chain(h http.Handler, hm ...Middleware) http.Handler {
	for i := len(hm) - 1; i >= 0; i-- {
		// 역순으로 미들웨어를 등록하면, 정순으로 실행된다.
		h = hm[i].handleMiddleware(h)
	}
	return h
}
