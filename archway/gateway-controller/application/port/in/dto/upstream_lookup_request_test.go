package dto

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUpStreamLookupRequestTrimsBearerPrefix(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/v1/member-api/api.example.com/api/users",
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
}

func TestNewUpStreamLookupRequestLeavesMissingTokenNil(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"http://unix/v1/upstream?path=/v1/member-api/api.example.com/api/users",
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
