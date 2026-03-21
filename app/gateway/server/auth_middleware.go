package server

import (
	"gateway/common/model/rewrite"
	"net/http"
)

type AuthMiddleware struct{}

func (a AuthMiddleware) handleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.Context().Value("targetURL").(*rewrite.RewitePathDTO)
		if targetURL.CheckAuthorization {
			// TODO 인증 로직 추가.
		}

		next.ServeHTTP(w, r)
	})
}
