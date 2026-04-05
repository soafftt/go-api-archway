package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	commonModel "gateway/common/model"
	"gateway/common/model/rewrite"
	"gateway/config"
	"gateway/gwe"
)

// --- 테스트 헬퍼 ---

func newTestConfig(baseURL string) *config.AppConfig {
	cfg := &config.AppConfig{}
	cfg.UpstreamLookup.BaseURL = baseURL
	cfg.HttpClient.MaxIdleConns = 100
	cfg.HttpClient.MaxIdleConnsPerHost = 100
	cfg.HttpClient.IdleConnTimeoutSeconds = 90
	cfg.HttpClient.TimeoutMilliSeconds = 5000
	return cfg
}

func newTestHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}
}

func newRewriteDTO() *rewrite.RewritePathDTO {
	return &rewrite.RewritePathDTO{
		Host:               "localhost:8080",
		Path:               "/api/v1/users",
		Method:             "GET",
		ResponseTimeout:    5000,
		RequestTimeout:     5000,
		CheckAuthorization: false,
		CacheTimeout:       0,
	}
}

// --- 단위 테스트 ---

func TestUpstreamLookupService_Lookup_Success(t *testing.T) {
	dto := newRewriteDTO()
	body, _ := json.Marshal(dto)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	cfg := newTestConfig(srv.URL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())

	result := svc.Lookup("/api/v1/users")

	if !result.Ok {
		t.Fatalf("Lookup 성공을 기대했으나 실패: %v", result.Error)
	}
	if result.Upstream.Host != dto.Host {
		t.Errorf("Host 불일치: got %s, want %s", result.Upstream.Host, dto.Host)
	}
	if result.Upstream.Path != dto.Path {
		t.Errorf("Path 불일치: got %s, want %s", result.Upstream.Path, dto.Path)
	}
}

func TestUpstreamLookupService_Lookup_NotFound(t *testing.T) {
	errResp := commonModel.ErrorResponse{
		Message: "NOT_FOUND_UPSTREAM_PATH",
		Detail:  "path not found",
	}
	body, _ := json.Marshal(errResp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	}))
	defer srv.Close()

	cfg := newTestConfig(srv.URL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())

	result := svc.Lookup("/not/exist")

	if result.Ok {
		t.Fatal("실패를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupUpstreamResult {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupUpstreamResult)
	}
}

func TestUpstreamLookupService_Lookup_TransportError(t *testing.T) {
	// 즉시 닫힌 서버로 transport 에러 유발
	noSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := noSrv.URL
	noSrv.Close()

	cfg := newTestConfig(baseURL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())
	result := svc.Lookup("/some/path")

	if result.Ok {
		t.Fatal("에러를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupTransport {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupTransport)
	}
}

func TestUpstreamLookupService_Lookup_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	cfg := newTestConfig(srv.URL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())

	result := svc.Lookup("/api/v1")

	if result.Ok {
		t.Fatal("실패를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupDecodeBody {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupDecodeBody)
	}
}

// --- 벤치마크 ---

func BenchmarkLookup_Success(b *testing.B) {
	dto := newRewriteDTO()
	body, _ := json.Marshal(dto)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	cfg := newTestConfig(srv.URL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := svc.Lookup("/api/v1/users")
		if !result.Ok {
			b.Fatalf("Lookup 실패: %v", result.Error)
		}
	}
}

func BenchmarkLookup_NotFound(b *testing.B) {
	errResp := commonModel.ErrorResponse{
		Message: "NOT_FOUND_UPSTREAM_PATH",
		Detail:  "path not found",
	}
	body, _ := json.Marshal(errResp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	}))
	defer srv.Close()

	cfg := newTestConfig(srv.URL + "?path=")
	svc := NewUpstreamLookupService(cfg, newTestHTTPClient())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		svc.Lookup("/not/exist")
	}
}

func BenchmarkBodyRead(b *testing.B) {
	body := []byte(`{"domain":"localhost:8080","path":"/api/v1/users","method":"GET","response_timeout":5000,"request_timeout":5000,"check_authorization":false,"cache_timeout":0}`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	client := newTestHTTPClient()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := bodyRead(resp); err != nil {
			b.Fatal(err)
		}
	}
}
