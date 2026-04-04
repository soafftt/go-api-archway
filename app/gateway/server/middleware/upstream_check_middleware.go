package middleware

import (
	"context"
	"gateway/common/code"
	gatewayContext "gateway/context"
	"gateway/gatewayerrors"
	"gateway/model"
	"gateway/server/response"
	"gateway/service"
	"net/http"
)

type UpstreamCheckMiddleware Middleware

type upstreamCheckMiddleware struct {
	upstreamLookupService service.UpstreamLookupService
}

func NewUpstreamCheckMiddleware(
	upstreamLookupService service.UpstreamLookupService,
) UpstreamCheckMiddleware {
	return &upstreamCheckMiddleware{
		upstreamLookupService: upstreamLookupService,
	}
}

func (m *upstreamCheckMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 현재 요청에 대해서 UpstreamLookupService 를 통해서 타겟 URL 을 조회.
		// 조회된 URL 을 ReverseProxy 의 Rewrite 함수에서 사용할 수 있도록 Context 에저장.
		lookupResult := m.upstreamLookupService.Lookup(r.URL.Path)
		if !lookupResult.Ok {
			var statusCode int
			var message string
			var detail string

			switch lookupResult.Error.Kind {
			case gatewayerrors.ErrLookupUpstreamResult:
				// common code 정의에 따라서 NOT_FOUND 시리즈는 404, 그 외는 500 처리.
				statusCode, message, detail = handleUpstreamResult(lookupResult)
			default:
				statusCode = http.StatusInternalServerError
				message = "INTERNAL_SERVER_ERROR"
				detail = "unknown error occurred"
			}

			response.HandErrorResponse(w, statusCode, message, detail)

			return
		}

		ctx := context.WithValue(r.Context(), gatewayContext.UpstreamContextKey, lookupResult.Upstream)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func handleUpstreamResult(lookupResult model.UpstreamLookupResult) (httpStatus int, message string, detail string) {
	switch lookupResult.Error.Message {
	case code.NOT_FOUND_UPSTREAM_SERVICE, code.NOT_FOUND_UPSTREAM_DOMAIN, code.NOT_FOUND_UPSTREAM_PATH:
		httpStatus = http.StatusNotFound
	default:
		httpStatus = http.StatusInternalServerError
	}

	return httpStatus, lookupResult.Error.Message, lookupResult.Error.Detail.Error()
}
