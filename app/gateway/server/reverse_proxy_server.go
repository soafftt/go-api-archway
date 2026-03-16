package server

import (
	"context"
	"gateway/common/model/rewrite"
	"gateway/service"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"strconv"
	"syscall"
)

type ReverseProxyServer struct {
	proxy   *httputil.ReverseProxy
	handler http.HandlerFunc
}

func NewReserveProxyServer(upstreamLookupService service.UpstreamLookupService) *ReverseProxyServer {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			targetURL := pr.In.Context().Value("targetURL").(*rewrite.RewitePathDTO)
			url := &url.URL{
				Scheme: "http",
				Host:   targetURL.Host,
				Path:   targetURL.Path,
			}

			pr.SetURL(url)

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
			// 에러 처리.
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 현재 요청에 대해서 UpstreamLookupService 를 통해서 타겟 URL 을 조회.
		// 조회된 URL 을 ReverseProxy 의 Rewrite 함수에서 사용할 수 있도록 Context 에 저장.
		path := r.URL.Path
		targetURL, err := upstreamLookupService.Lookup(path)

		if err != nil {
			http.Error(w, "Upstream lookup failed", http.StatusBadGateway)
			return
		}

		ctx := context.WithValue(r.Context(), "targetURL", &targetURL)
		r = r.WithContext(ctx)

		proxy.ServeHTTP(w, r)
	})

	return &ReverseProxyServer{
		proxy:   proxy,
		handler: handler,
	}

}

func (s *ReverseProxyServer) Start() {
	httpServe := http.Server{
		Addr:    ":80",
		Handler: s.handler,
	}

	go func() {
		func() {
			if r := recover(); r != nil {
				log.Printf("panic occurred: %v", r)
			}
		}()

		if err := httpServe.ListenAndServe(); err != nil {
			log.Fatalf("서버 실행 실패: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("서버 종료")
}
