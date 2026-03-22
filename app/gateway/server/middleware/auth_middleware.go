package middleware

import (
	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
	"net/http"
)

type AuthMiddleware Middleware

type authMiddleware struct{}

func NewAuthMiddleware() AuthMiddleware {
	return &authMiddleware{}
}

func (a *authMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)
		if targetURL.CheckAuthorization {
			// TODO 인증 로직 추가.
		}

		next.ServeHTTP(w, r)
	})
}
