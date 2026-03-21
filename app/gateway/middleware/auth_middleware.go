package middleware

import (
	"gateway/common/model/rewrite"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.Context().Value("targetURL").(*rewrite.RewitePathDTO)
		if targetURL.CheckAuthorization {
			// TODO 인증 로직 추가.
		}

		next.ServeHTTP(w, r)
	})
}
