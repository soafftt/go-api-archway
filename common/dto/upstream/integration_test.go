// author: copilot

package upstream

import (
	"encoding/json"
	"testing"
)

func TestUpStreamService_JSONUnmarshal(t *testing.T) {
	// 테스트 JSON 데이터
	jsonData := `{
		"servie": "user-service",
		"host": {
			"api.example.com": {
				"host": "upstream-server-1.internal:8080",
				"request": [
					{
						"path": "/api/users",
						"method": "GET",
						"requestTimeout": 5000,
						"responseTimeout": 10000,
						"checkAuthorization": true
					},
					{
						"path": "/api/users/{id}",
						"method": "GET",
						"requestTimeout": 3000,
						"responseTimeout": 5000,
						"checkAuthorization": true
					},
					{
						"path": "/api/users/{userId}/posts/{postId}",
						"method": "GET",
						"requestTimeout": 5000,
						"responseTimeout": 8000,
						"checkAuthorization": true
					}
				]
			},
			"admin.example.com": {
				"host": "admin-server.internal:8081",
				"request": [
					{
						"path": "/admin/users",
						"method": "GET",
						"requestTimeout": 10000,
						"responseTimeout": 15000,
						"checkAuthorization": true
					}
				]
			}
		}
	}`

	// JSON 파싱
	var service UpstreamService
	err := json.Unmarshal([]byte(jsonData), &service)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 기본 검증
	if service.Service != "user-service" {
		t.Errorf("Expected Service 'user-service', got '%s'", service.Service)
	}

	if len(service.HostMap) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(service.HostMap))
	}

	// api.example.com 검증
	apiHost := service.HostMap["api.example.com"]
	if apiHost == nil {
		t.Fatal("api.example.com not found")
	}

	if apiHost.Host != "upstream-server-1.internal:8080" {
		t.Errorf("Expected host 'upstream-server-1.internal:8080', got '%s'", apiHost.Host)
	}

	if len(apiHost.Request) != 3 {
		t.Errorf("Expected 3 requests for api.example.com, got %d", len(apiHost.Request))
	}

	// Router 초기화
	apiHost.InitializeRouter()

	// 라우팅 테스트
	tests := []struct {
		name         string
		path         string
		shouldFind   bool
		expectedPath string
	}{
		{"exact match", "/api/users", true, "/api/users"},
		{"path variable single", "/api/users/123", true, "/api/users/{id}"},
		{"path variable nested", "/api/users/123/posts/456", true, "/api/users/{userId}/posts/{postId}"},
		{"not found", "/api/products", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := apiHost.LookupPath(tt.path)

			if tt.shouldFind {
				if result == nil {
					t.Errorf("Expected to find path %s", tt.path)
					return
				}
				if result.Path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, result.Path)
				}
				if result.Method != "GET" {
					t.Errorf("Expected method GET, got %s", result.Method)
				}
			} else {
				if result != nil {
					t.Errorf("Expected not to find path %s, but got %v", tt.path, result)
				}
			}
		})
	}

	// admin.example.com 검증
	adminHost := service.HostMap["admin.example.com"]
	if adminHost == nil {
		t.Fatal("admin.example.com not found")
	}

	if adminHost.Host != "admin-server.internal:8081" {
		t.Errorf("Expected host 'admin-server.internal:8081', got '%s'", adminHost.Host)
	}

	if len(adminHost.Request) != 1 {
		t.Errorf("Expected 1 request for admin.example.com, got %d", len(adminHost.Request))
	}

	// LookupHost 테스트
	foundHost := service.LookupHost("api.example.com")
	if foundHost == nil {
		t.Error("LookupHost failed for api.example.com")
	}

	notFoundHost := service.LookupHost("nonexistent.example.com")
	if notFoundHost != nil {
		t.Error("LookupHost should return nil for non-existent host")
	}
}

func TestUpStreamService_ComplexScenario(t *testing.T) {
	jsonData := `{
		"servie": "product-service",
		"host": {
			"api.shop.com": {
				"host": "product-api.internal:8080",
				"request": [
					{
						"path": "/products",
						"method": "GET",
						"requestTimeout": 3000,
						"responseTimeout": 5000,
						"checkAuthorization": false
					},
					{
						"path": "/products/{productId}",
						"method": "GET",
						"requestTimeout": 2000,
						"responseTimeout": 4000,
						"checkAuthorization": false
					},
					{
						"path": "/products/{productId}/reviews",
						"method": "GET",
						"requestTimeout": 5000,
						"responseTimeout": 8000,
						"checkAuthorization": false
					},
					{
						"path": "/products/{productId}/reviews/{reviewId}",
						"method": "GET",
						"requestTimeout": 3000,
						"responseTimeout": 5000,
						"checkAuthorization": true
					},
					{
						"path": "/categories",
						"method": "GET",
						"requestTimeout": 2000,
						"responseTimeout": 3000,
						"checkAuthorization": false
					},
					{
						"path": "/categories/{categoryId}",
						"method": "GET",
						"requestTimeout": 2000,
						"responseTimeout": 3000,
						"checkAuthorization": false
					},
					{
						"path": "/categories/{categoryId}/products",
						"method": "GET",
						"requestTimeout": 5000,
						"responseTimeout": 8000,
						"checkAuthorization": false
					}
				]
			}
		}
	}`

	var service UpstreamService
	err := json.Unmarshal([]byte(jsonData), &service)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	host := service.LookupHost("api.shop.com")
	if host == nil {
		t.Fatal("Host not found")
	}

	host.InitializeRouter()

	// 복잡한 라우팅 시나리오 테스트
	tests := []struct {
		path              string
		expectedPattern   string
		expectedTimeout   int64
		expectedAuthCheck bool
	}{
		{"/products", "/products", 3000, false},
		{"/products/123", "/products/{productId}", 2000, false},
		{"/products/abc-def", "/products/{productId}", 2000, false},
		{"/products/123/reviews", "/products/{productId}/reviews", 5000, false},
		{"/products/123/reviews/456", "/products/{productId}/reviews/{reviewId}", 3000, true},
		{"/categories", "/categories", 2000, false},
		{"/categories/electronics", "/categories/{categoryId}", 2000, false},
		{"/categories/electronics/products", "/categories/{categoryId}/products", 5000, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := host.LookupPath(tt.path)

			if result == nil {
				t.Fatalf("Expected to find path %s", tt.path)
			}

			if result.Path != tt.expectedPattern {
				t.Errorf("Expected pattern %s, got %s", tt.expectedPattern, result.Path)
			}

			if result.RequestTimeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %d, got %d", tt.expectedTimeout, result.RequestTimeout)
			}

			if result.CheckAuthrozation != tt.expectedAuthCheck {
				t.Errorf("Expected auth check %v, got %v", tt.expectedAuthCheck, result.CheckAuthrozation)
			}
		})
	}
}

func TestUpStreamService_EmptyAndEdgeCases(t *testing.T) {
	t.Run("empty service", func(t *testing.T) {
		jsonData := `{
			"servie": "empty-service",
			"host": {}
		}`

		var service UpstreamService
		err := json.Unmarshal([]byte(jsonData), &service)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if service.Service != "empty-service" {
			t.Errorf("Expected 'empty-service', got '%s'", service.Service)
		}

		if len(service.HostMap) != 0 {
			t.Errorf("Expected 0 hosts, got %d", len(service.HostMap))
		}
	})

	t.Run("host with no requests", func(t *testing.T) {
		jsonData := `{
			"servie": "test-service",
			"host": {
				"test.com": {
					"host": "localhost:8080",
					"request": []
				}
			}
		}`

		var service UpstreamService
		err := json.Unmarshal([]byte(jsonData), &service)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		host := service.LookupHost("test.com")
		if host == nil {
			t.Fatal("Host not found")
		}

		host.InitializeRouter()

		result := host.LookupPath("/any/path")
		if result != nil {
			t.Error("Expected nil for non-existent path")
		}
	})
}
