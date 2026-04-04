package server

import (
	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
	"gateway/server/response"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/google/wire"
)

func NewGatewayReverseProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			targetURL := pr.In.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)
			url := &url.URL{
				Scheme: "http",
				Host:   targetURL.Host,
				Path:   targetURL.Path,
			}

			pr.SetURL(url)
			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
			// TODO x-forwarded-for 헤더 추가.
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
