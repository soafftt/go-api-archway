// author: copilot

package upstream

import (
	"encoding/json"
	"testing"
)

func TestUpStreamService_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"service_name": "member-api",
		"resources": [
			{
				"sub_domain": "api.example.com",
				"host": "upstream-server-1.internal:8080",
				"paths": [
					{
						"path": "/api/users",
						"method": "GET",
						"request_timeout": 5000,
						"response_timeout": 10000,
						"check_authorization": true,
						"cache_timeout": 10
					},
					{
						"path": "/api/users/{id}",
						"method": "GET",
						"request_timeout": 3000,
						"response_timeout": 5000,
						"check_authorization": true
					}
				]
			},
			{
				"sub_domain": "",
				"host": "default-user.internal:8081",
				"paths": [
					{
						"path": "/v1/member/",
						"method": "GET",
						"request_timeout": 5000,
						"response_timeout": 8000,
						"check_authorization": true
					}
				]
			}
		]
	}`

	var service UpstreamService
	err := json.Unmarshal([]byte(jsonData), &service)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	service.InitializeResourceIndex()

	if service.ServiceName != "member-api" {
		t.Errorf("Expected ServiceName 'member-api', got '%s'", service.ServiceName)
	}

	if len(service.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(service.Resources))
	}

	apiResource, isEmptyDomain := service.LookupResourceDomain("api.example.com")
	if apiResource == nil {
		t.Fatal("api.example.com not found")
	}
	if isEmptyDomain {
		t.Fatal("api.example.com should not use empty-domain fallback")
	}

	if apiResource.Host != "upstream-server-1.internal:8080" {
		t.Errorf("Expected host 'upstream-server-1.internal:8080', got '%s'", apiResource.Host)
	}

	if len(apiResource.Paths) != 2 {
		t.Errorf("Expected 2 paths for api.example.com, got %d", len(apiResource.Paths))
	}

	if apiResource.Paths[0].CacheTimeout != 10 {
		t.Errorf("Expected cache timeout 10, got %d", apiResource.Paths[0].CacheTimeout)
	}

	apiResource.InitializeRouter()
	result := apiResource.LookupPath("/api/users/123")
	if result == nil || result.Path != "/api/users/{id}" {
		t.Fatalf("Expected path '/api/users/{id}', got %#v", result)
	}

	fallbackResource, fallback := service.LookupResourceDomain("unknown.example.com")
	if fallbackResource == nil {
		t.Fatal("Expected fallback resource for empty subdomain")
	}
	if !fallback {
		t.Fatal("Expected fallback flag true")
	}
}

func TestUpStreamService_ComplexScenario(t *testing.T) {
	jsonData := `{
		"service_name": "product-api",
		"resources": [
			{
				"sub_domain": "api.shop.com",
				"host": "product-api.internal:8080",
				"paths": [
					{
						"path": "/products",
						"method": "GET",
						"request_timeout": 3000,
						"response_timeout": 5000,
						"check_authorization": false
					},
					{
						"path": "/products/{productId}",
						"method": "GET",
						"request_timeout": 2000,
						"response_timeout": 4000
					}
				]
			}
		]
	}`

	var service UpstreamService
	err := json.Unmarshal([]byte(jsonData), &service)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	service.InitializeResourceIndex()

	resource, isEmptyDomain := service.LookupResourceDomain("api.shop.com")
	if resource == nil {
		t.Fatal("Resource not found")
	}
	if isEmptyDomain {
		t.Fatal("api.shop.com should not be empty domain fallback")
	}

	resource.InitializeRouter()

	tests := []struct {
		path              string
		expectedPattern   string
		expectedTimeout   int64
		expectedAuthCheck bool
	}{
		{"/products", "/products", 3000, false},
		{"/products/123", "/products/{productId}", 2000, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := resource.LookupPath(tt.path)
			if result == nil {
				t.Fatalf("Expected to find path %s", tt.path)
			}
			if result.Path != tt.expectedPattern {
				t.Errorf("Expected pattern %s, got %s", tt.expectedPattern, result.Path)
			}
			if result.RequestTimeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %d, got %d", tt.expectedTimeout, result.RequestTimeout)
			}
			if result.CheckAuthorization != tt.expectedAuthCheck {
				t.Errorf("Expected auth check %v, got %v", tt.expectedAuthCheck, result.CheckAuthorization)
			}
		})
	}
}

func TestUpStreamService_EmptyAndEdgeCases(t *testing.T) {
	t.Run("empty service", func(t *testing.T) {
		jsonData := `{"service_name": "empty-api", "resources": []}`

		var service UpstreamService
		err := json.Unmarshal([]byte(jsonData), &service)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
		service.InitializeResourceIndex()

		if service.ServiceName != "empty-api" {
			t.Errorf("Expected 'empty-api', got '%s'", service.ServiceName)
		}

		if len(service.Resources) != 0 {
			t.Errorf("Expected 0 resources, got %d", len(service.Resources))
		}
	})

	t.Run("resource with no paths", func(t *testing.T) {
		jsonData := `{
			"service_name": "test-api",
			"resources": [
				{
					"sub_domain": "test.com",
					"host": "localhost:8080",
					"paths": []
				}
			]
		}`

		var service UpstreamService
		err := json.Unmarshal([]byte(jsonData), &service)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
		service.InitializeResourceIndex()

		resource, isEmptyDomain := service.LookupResourceDomain("test.com")
		if resource == nil {
			t.Fatal("Resource not found")
		}
		if isEmptyDomain {
			t.Fatal("test.com should not be empty domain fallback")
		}

		resource.InitializeRouter()
		result := resource.LookupPath("/any/path")
		if result != nil {
			t.Error("Expected nil for non-existent path")
		}
	})
}
