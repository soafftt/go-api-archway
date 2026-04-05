package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/common/model/rewrite"
	gatewayContext "gateway/context"
	"gateway/gwe"
	"gateway/model"
)

// --- mock UpstreamLookupService ---

type mockLookupService struct {
	result model.UpstreamLookupResult
}

func (m *mockLookupService) Lookup(_ string) model.UpstreamLookupResult {
	return m.result
}

func newSuccessLookupService() *mockLookupService {
	return &mockLookupService{
		result: model.NewUpstreamLookupResult(&rewrite.RewritePathDTO{
			Host:               "localhost:8080",
			Path:               "/api/v1/users",
			Method:             "GET",
			CheckAuthorization: false,
			CacheTimeout:       0,
		}),
	}
}

func newFailLookupService(kind gwe.LookupErrorKind, message string) *mockLookupService {
	return &mockLookupService{
		result: model.NewUpstreamLookupError(kind, message, nil),
	}
}

func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// --- 단위 테스트 ---

func TestUpstreamCheckMiddleware_Success(t *testing.T) {
	mw := NewUpstreamCheckMiddleware(newSuccessLookupService())
	handler := mw.HandleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstream := r.Context().Value(gatewayContext.UpstreamContextKey)
		if upstream == nil {
			t.Error("컨텍스트에 upstream이 없음")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("상태 코드 불일치: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUpstreamCheckMiddleware_NotFound(t *testing.T) {
	svc := newFailLookupService(gwe.ErrLookupUpstreamResult, "NOT_FOUND_UPSTREAM_PATH")
	// NOT_FOUND_UPSTREAM_PATH 케이스는 detail error 가 필요하므로 직접 구성
	svc.result.Error.Detail = errWrapper("path not found")
	mw := NewUpstreamCheckMiddleware(svc)
	handler := mw.HandleMiddleware(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/not/exist", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("상태 코드 불일치: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpstreamCheckMiddleware_InternalError(t *testing.T) {
	svc := newFailLookupService(gwe.ErrLookupTransport, gwe.ErrMsgTransport)
	svc.result.Error.Detail = errWrapper("transport error")
	mw := NewUpstreamCheckMiddleware(svc)
	handler := mw.HandleMiddleware(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("상태 코드 불일치: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUpstreamCheckMiddleware_ContextPropagation(t *testing.T) {
	expected := &rewrite.RewritePathDTO{
		Host: "upstream-host:9090",
		Path: "/v2/resource",
	}
	svc := &mockLookupService{result: model.NewUpstreamLookupResult(expected)}
	mw := NewUpstreamCheckMiddleware(svc)

	var gotUpstream *rewrite.RewritePathDTO
	handler := mw.HandleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUpstream = r.Context().Value(gatewayContext.UpstreamContextKey).(*rewrite.RewritePathDTO)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v2/resource", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotUpstream == nil || gotUpstream.Host != expected.Host {
		t.Errorf("컨텍스트 upstream 불일치: got %v, want %v", gotUpstream, expected)
	}
}

// errWrapper - 에러 인터페이스 구현 헬퍼
type errWrapper string

func (e errWrapper) Error() string { return string(e) }

// --- 벤치마크 ---

func BenchmarkUpstreamCheckMiddleware_Success(b *testing.B) {
	mw := NewUpstreamCheckMiddleware(newSuccessLookupService())
	handler := mw.HandleMiddleware(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkUpstreamCheckMiddleware_NotFound(b *testing.B) {
	svc := newFailLookupService(gwe.ErrLookupUpstreamResult, "NOT_FOUND_UPSTREAM_PATH")
	svc.result.Error.Detail = errWrapper("path not found")
	mw := NewUpstreamCheckMiddleware(svc)
	handler := mw.HandleMiddleware(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/not/exist", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkUpstreamCheckMiddleware_ContextWithValue(b *testing.B) {
	// context.WithValue 비용 단독 측정
	dto := &rewrite.RewritePathDTO{Host: "localhost:8080", Path: "/api/v1"}
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.WithValue(req.Context(), gatewayContext.UpstreamContextKey, dto)
		_ = req.WithContext(ctx)
	}
}

func BenchmarkHandleUpstreamResult(b *testing.B) {
	cases := []struct {
		name   string
		result model.UpstreamLookupResult
	}{
		{
			name: "NotFoundPath",
			result: func() model.UpstreamLookupResult {
				r := model.NewUpstreamLookupError(gwe.ErrLookupUpstreamResult, "NOT_FOUND_UPSTREAM_PATH", errWrapper("path not found"))
				return r
			}(),
		},
		{
			name: "InternalError",
			result: func() model.UpstreamLookupResult {
				r := model.NewUpstreamLookupError(gwe.ErrLookupUpstreamResult, "INTERNAL_ERROR", errWrapper("internal"))
				return r
			}(),
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _ = handleUpstreamResult(tc.result)
			}
		})
	}
}

// JSON 직렬화 비용 측정 (에러 응답)
func BenchmarkErrorResponseJSON(b *testing.B) {
	type errResp struct {
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}
	e := errResp{Message: "NOT_FOUND_UPSTREAM_PATH", Detail: "path not found"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(e)
	}
}
