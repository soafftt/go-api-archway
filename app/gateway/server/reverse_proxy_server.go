package server

import (
	"context"
	"encoding/json"
	commonCode "gateway/common/code"
	"gateway/common/model/rewrite"
	"gateway/middleware"
	gatewaymodel "gateway/model"
	"gateway/service"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/google/wire"
)

const (
	REVERS_PROXY_CONTEXT_KEY = "upstream"
)

type ReverseProxyServer struct {
	proxy                 *httputil.ReverseProxy
	upstreamLookupService service.UpstreamLookupService
}

func NewReserveProxyServer(upstreamLookupService service.UpstreamLookupService) *ReverseProxyServer {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			targetURL := pr.In.Context().Value(REVERS_PROXY_CONTEXT_KEY).(*rewrite.RewitePathDTO)
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
			targetURL := res.Request.Context().Value("targetURL").(*rewrite.RewitePathDTO)
			if targetURL.CacheTimeout > 0 {
				res.Header.Set("Cache-Control", "max-age="+strconv.FormatInt(targetURL.CacheTimeout, 10))
			}

			// 캐시처리 등등등.
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// TODO: Logging
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		},
	}

	return &ReverseProxyServer{
		proxy:                 proxy,
		upstreamLookupService: upstreamLookupService,
	}

}

var ReverseProxySet = wire.NewSet(
	NewReserveProxyServer,
)

func writeErrorJSON(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(gatewaymodel.NewErrorResponse(code, message))
}

func (s *ReverseProxyServer) Start() {
	// upstreamCheck 는 Service 를 주입 받아야 하기 때문에 이와 같이 처리.
	// 이걸 여기에다 둘수 밖에 없을까? 고민이 필요할듯.
	upstreamCheckMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 현재 요청에 대해서 UpstreamLookupService 를 통해서 타겟 URL 을 조회.
			// 조회된 URL 을 ReverseProxy 의 Rewrite 함수에서 사용할 수 있도록 Context 에저장.
			lookupResult := s.upstreamLookupService.Lookup(r.URL.Path)
			if !lookupResult.Ok {
				if strings.HasPrefix(lookupResult.Error.Message, "UNIX_SOCKET_") {
					writeErrorJSON(w, http.StatusInternalServerError, "unkonwn_error", "unknown error occurred")
				} else {
					// comonCode 정의 코드에 따라서 NOT_FOUND 시리즈는 404, 그 외는 500 처리.
					var statusCode int
					switch lookupResult.Error.Message {
					case commonCode.NOT_FOUND_UPSTERAM_SERVICE, commonCode.NOT_FOUND_UPSTERAM_DOMAIN, commonCode.NOT_FOUND_UPSTERMA_PATH:
						statusCode = http.StatusNotFound
					default:
						statusCode = http.StatusInternalServerError
					}

					writeErrorJSON(w, statusCode, lookupResult.Error.Message, "")
				}
				return
			}

			ctx := context.WithValue(r.Context(), REVERS_PROXY_CONTEXT_KEY, lookupResult.Upstream)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}

	httpServe := http.Server{
		Addr: ":80",
		Handler: middleware.Chain(
			s.proxy,
			upstreamCheckMiddleware,
		),
	}

	go func() {
		func() {
			if r := recover(); r != nil {
				log.Printf("panic occurred: %v", r)
			}
		}()

		log.Printf("Start Server: %s", httpServe.Addr)

		if err := httpServe.ListenAndServe(); err != nil {
			log.Fatalf("서버 실행 실패: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("서버 종료")
}
