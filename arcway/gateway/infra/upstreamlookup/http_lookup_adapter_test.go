package upstreamlookup

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
		Host:            "localhost:8080",
		Path:            "/api/v1/users",
		Method:          "GET",
		ResponseTimeout: 5000,
		RequestTimeout:  5000,
		CacheTimeout:    0,
	}
}

func TestHTTPUpstreamLookupAdapter_Lookup_Success(t *testing.T) {
	dto := newRewriteDTO()
	body, _ := json.Marshal(dto)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	adapter := NewHTTPUpstreamLookupAdapter(newTestConfig(srv.URL+"?path="), newTestHTTPClient())
	result := adapter.Lookup("/api/v1/users")

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

func TestHTTPUpstreamLookupAdapter_Lookup_NotFound(t *testing.T) {
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

	adapter := NewHTTPUpstreamLookupAdapter(newTestConfig(srv.URL+"?path="), newTestHTTPClient())
	result := adapter.Lookup("/not/exist")

	if result.Ok {
		t.Fatal("실패를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupUpstreamResult {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupUpstreamResult)
	}
}

func TestHTTPUpstreamLookupAdapter_Lookup_TransportError(t *testing.T) {
	noSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := noSrv.URL
	noSrv.Close()

	adapter := NewHTTPUpstreamLookupAdapter(newTestConfig(baseURL+"?path="), newTestHTTPClient())
	result := adapter.Lookup("/some/path")

	if result.Ok {
		t.Fatal("에러를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupTransport {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupTransport)
	}
}

func TestHTTPUpstreamLookupAdapter_Lookup_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	adapter := NewHTTPUpstreamLookupAdapter(newTestConfig(srv.URL+"?path="), newTestHTTPClient())
	result := adapter.Lookup("/api/v1")

	if result.Ok {
		t.Fatal("실패를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupDecodeBody {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupDecodeBody)
	}
}
