package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
	"gateway/gwe"
	"gateway/model"
)

// 테스트용 단순 리버스 프록시 생성
func newTestProxy(upstreamAddr string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			dto := pr.In.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)
			u := &url.URL{
				Scheme: "http",
				Host:   upstreamAddr,
				Path:   dto.Path,
			}
			pr.SetURL(u)
			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
		},
	}
}

// --- 전체 미들웨어 체인 단위 테스트 ---

func TestMiddlewareChain_FullPipeline_Success(t *testing.T) {
	upstreamCalled := false
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	dto := &rewrite.RewritePathDTO{
		Host:               upstreamSrv.Listener.Addr().String(),
		Path:               "/api/v1/users",
		CheckAuthorization: false,
	}
	svc := &mockLookupService{result: model.NewUpstreamLookupResult(dto)}

	upstreamMW := NewUpstreamCheckMiddleware(svc)
	proxy := newTestProxy(upstreamSrv.Listener.Addr().String())

	handler := Chain(proxy, upstreamMW)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !upstreamCalled {
		t.Error("upstream 핸들러가 호출되지 않음")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("상태 코드 불일치: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMiddlewareChain_FullPipeline_NotFound(t *testing.T) {
	svc := newFailLookupService(gwe.ErrLookupUpstreamResult, "NOT_FOUND_UPSTREAM_PATH")
	svc.result.Error.Detail = errWrapper("path not found")

	upstreamMW := NewUpstreamCheckMiddleware(svc)
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("proxy가 호출되어서는 안 됨")
	})

	handler := Chain(proxy, upstreamMW)

	req := httptest.NewRequest(http.MethodGet, "/not/exist", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("상태 코드 불일치: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// --- 전체 파이프라인 벤치마크 ---

func BenchmarkFullPipeline_WithUpstream(b *testing.B) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	dto := &rewrite.RewritePathDTO{
		Host:               upstreamSrv.Listener.Addr().String(),
		Path:               "/api/v1/users",
		CheckAuthorization: false,
		CacheTimeout:       0,
	}
	svc := &mockLookupService{result: model.NewUpstreamLookupResult(dto)}

	upstreamMW := NewUpstreamCheckMiddleware(svc)
	proxy := newTestProxy(upstreamSrv.Listener.Addr().String())
	handler := Chain(proxy, upstreamMW)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkFullPipeline_MockProxy(b *testing.B) {
	// 실제 네트워크 비용 제거 후 미들웨어 체인 순수 비용 측정
	dto := &rewrite.RewritePathDTO{
		Host:               "localhost:8080",
		Path:               "/api/v1/users",
		CheckAuthorization: false,
		CacheTimeout:       0,
	}
	svc := &mockLookupService{result: model.NewUpstreamLookupResult(dto)}

	upstreamMW := NewUpstreamCheckMiddleware(svc)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Chain(finalHandler, upstreamMW)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkContextChain(b *testing.B) {
	// context.WithValue + r.WithContext 체인 비용 측정
	dto := &rewrite.RewritePathDTO{Host: "localhost:8080", Path: "/api/v1"}
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto)
		r2 := req.WithContext(ctx)
		_ = r2.Context().Value(gatewayContext.UpstreamContextKey)
	}
}

func BenchmarkChain_SingleMiddleware(b *testing.B) {
	// Chain 함수 + 단일 미들웨어 오버헤드 측정
	dto := &rewrite.RewritePathDTO{
		Host:               "localhost:8080",
		Path:               "/api/v1/users",
		CheckAuthorization: false,
	}
	svc := &mockLookupService{result: model.NewUpstreamLookupResult(dto)}

	upstreamMW := NewUpstreamCheckMiddleware(svc)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := Chain(finalHandler, upstreamMW)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

