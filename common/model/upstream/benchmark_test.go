package upstream

import (
	"encoding/json"
	"testing"
)

var benchmarkServiceJSON = []byte(`{
	"service_name": "member-api",
	"resouces": [
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
}`)

func mustBuildBenchmarkService(b *testing.B) *UpstreamService {
	b.Helper()

	var service UpstreamService
	if err := json.Unmarshal(benchmarkServiceJSON, &service); err != nil {
		b.Fatalf("unmarshal failed: %v", err)
	}
	service.InitializeResourceIndex()

	return &service
}

func BenchmarkUpstreamService_UnmarshalAndInitialize(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var service UpstreamService
		if err := json.Unmarshal(benchmarkServiceJSON, &service); err != nil {
			b.Fatalf("unmarshal failed: %v", err)
		}
		service.InitializeResourceIndex()
	}
}

func BenchmarkUpstreamService_LookupResourceDomain(b *testing.B) {
	b.ReportAllocs()

	service := mustBuildBenchmarkService(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resource, _ := service.LookupResourceDomain("api.example.com")
		if resource == nil {
			b.Fatal("resource not found")
		}
	}
}

func BenchmarkUpstreamResource_LookupPath(b *testing.B) {
	b.ReportAllocs()

	service := mustBuildBenchmarkService(b)
	resource, _ := service.LookupResourceDomain("api.example.com")
	if resource == nil {
		b.Fatal("resource not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := resource.LookupPath("/api/users/123")
		if result == nil {
			b.Fatal("path not found")
		}
	}
}
