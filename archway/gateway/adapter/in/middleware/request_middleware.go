package middleware

import (
	"context"
	in2 "core/adapter/in"
	"core/errs"
	"encoding/json"
	"gateway/adapter/in/ctxkey"
	"gateway/application/port/in"
	"net/http"
	"strings"
)

type RequestMiddleware Middleware

type requestMiddleware struct {
	upstreamLookupUseCase in.UpstreamLookupUseCase
}

func NewRequestMiddleware(
	upstreamLookupUseCase in.UpstreamLookupUseCase,
) RequestMiddleware {
	return requestMiddleware{
		upstreamLookupUseCase: upstreamLookupUseCase,
	}
}

func (r requestMiddleware) HandleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		jsonEncoder := json.NewEncoder(writer)

		target := request.URL.Path
		if target == "/" {
			writer.WriteHeader(http.StatusNotFound)
			writer.Header().Set("Content-Type", "application/json")
			_ = jsonEncoder.Encode(in2.NewErrorResponse(errs.ERR_INVALID_TARGET))
			return
		}

		authorization := request.Header.Get("Authorization")
		authorization = strings.TrimPrefix(authorization, "Bearer ")

		accessToken := &authorization
		if authorization == "" {
			accessToken = nil
		}

		lookupResult, err := r.upstreamLookupUseCase.Lookup(target, accessToken)
		// toto: handle ErrorResponse 같은 것을 만들면 좋을듯
		if err != nil {
			// 에러처리.
			switch errs.ToArchwayFromError(err) {
			case errs.ERR_INVALID_TARGET:
				writer.WriteHeader(http.StatusBadRequest)
			case errs.ERR_NOT_FOUND_DOMAIN_RESOURCE:
			case errs.ERR_NOT_FOUND_DOMAIN_RESROUCE_PATH:
			case errs.ERR_NOT_FOUND_SERVICE:
				writer.WriteHeader(http.StatusNotFound)
			default:
				writer.WriteHeader(http.StatusInternalServerError)
			}

			writer.Header().Set("Content-Type", "application/json")
			_ = jsonEncoder.Encode(in2.NewErrorResponse(err))
			return
		}

		// 필수값 할당
		ctx := request.Context()
		ctx = context.WithValue(ctx, ctxkey.LOOKUP_KEY, lookupResult)

		// request 에 context 를 덮어씀.
		request = request.WithContext(ctx)

		next.ServeHTTP(writer, request)
	})
}
