package dto

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUpStreamLookupRequestTrimsBearerPrefix(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/member/v1/user/info",
		nil,
	)
	request.Header.Set("Authorization", "Bearer access-token")

	result, err := NewUpStreamLookupRequest(request)
	if err != nil {
		t.Fatalf("expected lookup request to parse: %v", err)
	}
	if result.AccessToken == nil {
		t.Fatal("expected access token")
	}
	if *result.AccessToken != "access-token" {
		t.Fatalf("expected access token %q, got %q", "access-token", *result.AccessToken)
	}
	if result.Service != "member" || result.Version != "v1" || result.Domain != "user" || result.Path != "user/info" {
		t.Fatalf("unexpected lookup request: %+v", result)
	}
}

func TestNewUpStreamLookupRequestLeavesMissingTokenNil(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/member/v1/user/info",
		nil,
	)

	result, err := NewUpStreamLookupRequest(request)
	if err != nil {
		t.Fatalf("expected lookup request to parse: %v", err)
	}
	if result.AccessToken != nil {
		t.Fatalf("expected nil access token, got %q", *result.AccessToken)
	}
}

func TestNewUpStreamLookupRequestSupportsLegacyVersionFirstPath(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/v1/member/user/info",
		nil,
	)

	result, err := NewUpStreamLookupRequest(request)
	if err != nil {
		t.Fatalf("expected lookup request to parse: %v", err)
	}
	if result.Service != "member" || result.Version != "v1" || result.Domain != "user" || result.Path != "user/info" {
		t.Fatalf("unexpected lookup request: %+v", result)
	}
}

func TestNewUpStreamLookupRequestRejectsIncompletePath(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/member/v1/user",
		nil,
	)

	if _, err := NewUpStreamLookupRequest(request); err == nil {
		t.Fatal("expected incomplete gateway path to return an error")
	}
}
