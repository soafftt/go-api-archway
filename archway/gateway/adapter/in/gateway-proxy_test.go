package in

import (
	"context"
	applicationIn "gateway/application/port/in"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGatewayProxyRewritesUpstreamHostAndPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/echo/ping" {
			t.Errorf("expected upstream path %q, got %q", "/echo/ping", request.URL.Path)
		}
		_, _ = writer.Write([]byte("upstream-ok"))
	}))
	defer upstream.Close()

	lookupResult := applicationIn.UpstreamLookupResult{
		Host:   upstream.URL,
		Path:   "/echo/ping",
		Method: http.MethodGet,
	}
	request := httptest.NewRequest(http.MethodGet, "http://gateway.local/v1/e2e-service/echo.local/echo/ping", nil)
	request = request.WithContext(context.WithValue(request.Context(), UpstreamLookupKey, lookupResult))
	response := httptest.NewRecorder()

	newReversProxy().ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read proxy response: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, result.StatusCode, body)
	}
	if string(body) != "upstream-ok" {
		t.Fatalf("expected upstream response %q, got %q", "upstream-ok", body)
	}
}
