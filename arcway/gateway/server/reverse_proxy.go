package server

import (
	"net/http"
	"net/http/httputil"
	"strconv"

	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
	"gateway/server/response"

	"github.com/google/wire"
)

// Reverse Proxy 처리.
func NewGatewayReverseProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			targetURL := pr.In.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)

			// SetURL은 target.Path + incoming.Path 를 join하므로 Out.URL 을 직접 설정
			pr.Out.URL.Scheme = "http"
			pr.Out.URL.Host = targetURL.Host
			pr.Out.URL.Path = targetURL.Path
			pr.Out.URL.RawPath = ""
			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)

			if targetURL.UserKey != "" {
				pr.Out.Header.Set("X-UserId", targetURL.UserKey.(string))
			}
		},
		ModifyResponse: func(res *http.Response) error {
			targetURL := res.Request.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)
			if targetURL.CacheTimeout > 0 {
				res.Header.Set("Cache-Control", "max-age="+strconv.FormatInt(targetURL.CacheTimeout, 10))
			}

			// 캐시처리 등등등.
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// TODO: Logging
			response.HandErrorResponse(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		},
	}
}

var ReverseProxySet = wire.NewSet(
	NewGatewayReverseProxy,
)
