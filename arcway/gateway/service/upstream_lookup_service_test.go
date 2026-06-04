package service

import (
	"testing"

	"gateway/common/model/rewrite"
	"gateway/gwe"
	"gateway/model"
)

type mockUpstreamLookupPort struct {
	result model.UpstreamLookupResult
}

func (m *mockUpstreamLookupPort) Lookup(_ string) model.UpstreamLookupResult {
	return m.result
}

func TestUpstreamLookupService_Lookup_Success(t *testing.T) {
	expected := model.NewUpstreamLookupResult(&rewrite.RewritePathDTO{
		Host: "localhost:8080",
		Path: "/api/v1/users",
	})
	svc := NewUpstreamLookupService(&mockUpstreamLookupPort{result: expected})

	result := svc.Lookup("/api/v1/users")

	if !result.Ok {
		t.Fatalf("Lookup 성공을 기대했으나 실패: %v", result.Error)
	}
	if result.Upstream.Host != expected.Upstream.Host {
		t.Errorf("Host 불일치: got %s, want %s", result.Upstream.Host, expected.Upstream.Host)
	}
}

func TestUpstreamLookupService_Lookup_Error(t *testing.T) {
	expected := model.NewUpstreamLookupError(gwe.ErrLookupTransport, gwe.ErrMsgTransport, errWrapper("transport error"))
	svc := NewUpstreamLookupService(&mockUpstreamLookupPort{result: expected})

	result := svc.Lookup("/api/v1/users")

	if result.Ok {
		t.Fatal("실패를 기대했으나 성공")
	}
	if result.Error.Kind != gwe.ErrLookupTransport {
		t.Errorf("에러 Kind 불일치: got %s, want %s", result.Error.Kind, gwe.ErrLookupTransport)
	}
}

type errWrapper string

func (e errWrapper) Error() string { return string(e) }
