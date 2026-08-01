package middleware

import (
	"context"
	coreAdapterIn "core/adapter/in"
	"core/errs"
	"encoding/json"
	"fmt"
	"gateway/adapter/in/ctxkey"
	portIn "gateway/application/port/in"
	portOutRateLimit "gateway/application/port/out/ratelimiter"
	"net/http"
	"strings"
)

type RequestMiddleware Middleware

type requestMiddleware struct {
	upstreamLookupUseCase portIn.UpstreamLookupUseCase
	rateLimiterPort       portOutRateLimit.RateLimiterPort
}

func NewRequestMiddleware(
	upstreamLookupUseCase portIn.UpstreamLookupUseCase,
	rateLimiterPort portOutRateLimit.RateLimiterPort,
) RequestMiddleware {
	return &requestMiddleware{
		upstreamLookupUseCase: upstreamLookupUseCase,
		rateLimiterPort:       rateLimiterPort,
	}
}

func (r requestMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		jsonEncoder := json.NewEncoder(writer)

		target := request.URL.Path
		if target == "/" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound)
			_ = jsonEncoder.Encode(coreAdapterIn.NewErrorResponse(errs.ERR_INVALID_TARGET))
			return
		}

		authorization := request.Header.Get("Authorization")
		authorization = strings.TrimPrefix(authorization, "Bearer ")

		accessToken := &authorization
		if authorization == "" {
			accessToken = nil
		}

		lookupResult, err := r.upstreamLookupUseCase.Lookup(target, accessToken, portIn.UnixGrpc)
		// toto: handle ErrorResponse 같은 것을 만들면 좋을듯
		if err != nil {
			// 에러처리.
			statusCode := http.StatusInternalServerError
			switch errs.ToArchwayFromError(err) {
			case errs.ERR_INVALID_TARGET:
				statusCode = http.StatusBadRequest
			case errs.ERR_NOT_FOUND_DOMAIN_RESOURCE:
			case errs.ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH:
			case errs.ERR_NOT_FOUND_SERVICE:
				statusCode = http.StatusNotFound
			}

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(statusCode)
			_ = jsonEncoder.Encode(coreAdapterIn.NewErrorResponse(err))
			return
		}

		// ratelimit 체크
		serviceName := lookupResult.ServiceName
		originalPath := lookupResult.OriginPath
		rateLimitCount := lookupResult.RateLimitCount

		if allow := r.chekAllowRateLimitWithErrorResponse(writer, jsonEncoder, serviceName, originalPath, rateLimitCount); !allow {
			return
		}

		// 필수값 할당
		ctx := request.Context()
		ctx = context.WithValue(ctx, ctxkey.UpstreamLookupKey, lookupResult)

		// request 에 context 를 덮어씀.
		request = request.WithContext(ctx)

		next.ServeHTTP(writer, request)
	})
}

func (r requestMiddleware) chekAllowRateLimitWithErrorResponse(writer http.ResponseWriter, encoder *json.Encoder, serviceName string, originPath string, rateLimitCount int64) bool {
	if rateLimitCount <= 0 {
		return true
	}

	if allow, remainToken, remainTokenPerSecond := r.rateLimiterPort.Allow(serviceName, originPath, int32(rateLimitCount)); !allow {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", remainToken))
		// 기존 오탈자 헤더와 호환성 유지를 위해 두 값을 함께 기록한다.
		writer.Header().Set("X-RateLimit-Remaining-PerSeconds", fmt.Sprintf("%d", remainTokenPerSecond))
		writer.Header().Set("X-RateLimit-Remaining-PerSeonds", fmt.Sprintf("%d", remainTokenPerSecond))
		writer.WriteHeader(http.StatusTooManyRequests)
		_ = encoder.Encode(coreAdapterIn.NewErrorResponse(errs.ERR_TOO_MANY_REQUESTS))
		return false
	}

	return true
}
