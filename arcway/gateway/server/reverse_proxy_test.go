package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
)

// --- 단위 테스트 ---

func TestReverseProxy_Rewrite_SetsUpstreamURL(t *testing.T) {
	upstreamCalled := false
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		// Out.URL 직접 설정 방식으로 path가 완전 대체됨을 검증
		if r.URL.Path != "/api/v1/users" {
			t.Errorf("upstream path 불일치: got %s, want /api/v1/users", r.URL.Path)
		}
		if r.Header.Get("X-Forwarded-Host") == "" {
			t.Error("X-Forwarded-Host 헤더 없음")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	proxy := NewGatewayReverseProxy()

	dto := &rewrite.RewritePathDTO{
		Host:         upstreamSrv.Listener.Addr().String(),
		Path:         "/api/v1/users",
		CacheTimeout: 0,
	}

	// incoming path와 target path가 다른 경우 — path 완전 대체 검증
	req := httptest.NewRequest(http.MethodGet, "/gateway/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto))
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if !upstreamCalled {
		t.Error("upstream 핸들러가 호출되지 않음")
	}
}

func TestReverseProxy_ModifyResponse_CacheHeader(t *testing.T) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	proxy := NewGatewayReverseProxy()

	dto := &rewrite.RewritePathDTO{
		Host:         upstreamSrv.Listener.Addr().String(),
		Path:         "/api/v1/cached",
		CacheTimeout: 300,
	}

	req := httptest.NewRequest(http.MethodGet, "/gateway/cached", nil)
	req = req.WithContext(context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto))
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	cacheHeader := rec.Header().Get("Cache-Control")
	if cacheHeader != "max-age=300" {
		t.Errorf("Cache-Control 헤더 불일치: got %q, want %q", cacheHeader, "max-age=300")
	}
}

func TestReverseProxy_ModifyResponse_NoCacheHeader_WhenZero(t *testing.T) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	proxy := NewGatewayReverseProxy()

	dto := &rewrite.RewritePathDTO{
		Host:         upstreamSrv.Listener.Addr().String(),
		Path:         "/api/v1/nocache",
		CacheTimeout: 0,
	}

	req := httptest.NewRequest(http.MethodGet, "/gateway/nocache", nil)
	req = req.WithContext(context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto))
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if cc := rec.Header().Get("Cache-Control"); cc != "" {
		t.Errorf("Cache-Control 헤더가 있어서는 안 됨: got %q", cc)
	}
}

// --- 벤치마크 ---

func BenchmarkReverseProxy_Rewrite(b *testing.B) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	proxy := NewGatewayReverseProxy()

	dto := &rewrite.RewritePathDTO{
		Host:         upstreamSrv.Listener.Addr().String(),
		Path:         "/api/v1/users",
		CacheTimeout: 0,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/gateway/users", nil)
		req = req.WithContext(context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto))
		rec := httptest.NewRecorder()
		proxy.ServeHTTP(rec, req)
	}
}

func BenchmarkReverseProxy_Rewrite_WithCache(b *testing.B) {
	upstreamSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstreamSrv.Close()

	proxy := NewGatewayReverseProxy()

	dto := &rewrite.RewritePathDTO{
		Host:         upstreamSrv.Listener.Addr().String(),
		Path:         "/api/v1/cached",
		CacheTimeout: 300,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/gateway/cached", nil)
		req = req.WithContext(context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto))
		rec := httptest.NewRecorder()
		proxy.ServeHTTP(rec, req)
	}
}

func BenchmarkNewGatewayReverseProxy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewGatewayReverseProxy()
	}
}
