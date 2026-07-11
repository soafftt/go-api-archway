package utils

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestToStructDecodesResponseBody(t *testing.T) {
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"name":"archway"}`)),
	}

	result, err := ToStruct[struct {
		Name string `json:"name"`
	}](response)
	if err != nil {
		t.Fatalf("expected response body to decode: %v", err)
	}
	if result.Name != "archway" {
		t.Fatalf("expected name %q, got %q", "archway", result.Name)
	}
}

func TestToStructReturnsDecodeError(t *testing.T) {
	response := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"name":`)),
	}

	if _, err := ToStruct[map[string]string](response); err == nil {
		t.Fatal("expected malformed response body to return an error")
	}
}
