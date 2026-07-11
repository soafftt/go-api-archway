package cache

import (
	coreUtils "core/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
)

const parseBenchmarkHS256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`

type parseBenchmarkDataset struct {
	serviceCount   int
	subdomains     int
	pathsPerDomain int
	withAuth       bool
}

func BenchmarkParseToUpstreamServiceWithInitializeScaled(b *testing.B) {
	cases := []struct {
		name    string
		dataset parseBenchmarkDataset
	}{
		{
			name: "10-services-x-4-subdomains-x-16-paths",
			dataset: parseBenchmarkDataset{
				serviceCount:   10,
				subdomains:     4,
				pathsPerDomain: 16,
				withAuth:       true,
			},
		},
		{
			name: "50-services-x-8-subdomains-x-32-paths",
			dataset: parseBenchmarkDataset{
				serviceCount:   50,
				subdomains:     8,
				pathsPerDomain: 32,
				withAuth:       true,
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			requests := buildParseBenchmarkRequests("parse-bench", tc.dataset)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := coreUtils.ParseToUpstreamServiceWithInitialize(requests)
				if len(results) != tc.dataset.serviceCount {
					b.Fatalf("expected %d services, got %d", tc.dataset.serviceCount, len(results))
				}
			}
		})
	}
}

func buildParseBenchmarkRequests(prefix string, dataset parseBenchmarkDataset) []coreUtils.ParseRequest {
	requests := make([]coreUtils.ParseRequest, 0, dataset.serviceCount)
	for serviceIndex := 0; serviceIndex < dataset.serviceCount; serviceIndex++ {
		serviceName := fmt.Sprintf("%s-%d", prefix, serviceIndex)
		requests = append(requests, coreUtils.NewParseRequest(serviceName, buildParseBenchmarkServiceJSON(serviceName, dataset)))
	}
	return requests
}

func buildParseBenchmarkServiceJSON(serviceName string, dataset parseBenchmarkDataset) string {
	payload := map[string]any{
		"service_name": serviceName,
	}

	if dataset.withAuth {
		payload["authorization"] = map[string]any{
			"algorithm": "HS256",
			"key_data":  base64.StdEncoding.EncodeToString([]byte(parseBenchmarkHS256JWKJSON)),
			"user_key":  "user_id",
		}
	}

	resources := make([]map[string]any, 0, dataset.subdomains)
	for subdomainIndex := 0; subdomainIndex < dataset.subdomains; subdomainIndex++ {
		paths := make([]map[string]any, 0, dataset.pathsPerDomain)
		for pathIndex := 0; pathIndex < dataset.pathsPerDomain; pathIndex++ {
			path := fmt.Sprintf("/api/%d/items/%d", subdomainIndex, pathIndex)
			if pathIndex%3 == 0 {
				path = fmt.Sprintf("/api/%d/users/{userId}/orders/%d", subdomainIndex, pathIndex)
			}

			paths = append(paths, map[string]any{
				"path":                path,
				"method":              "GET",
				"request_timeout":     3000 + pathIndex,
				"response_timeout":    5000 + pathIndex,
				"check_authorization": dataset.withAuth && subdomainIndex == 0 && pathIndex == dataset.pathsPerDomain-1,
				"cache_timeout":       pathIndex % 10,
			})
		}

		resources = append(resources, map[string]any{
			"sub_domain": fmt.Sprintf("api-%d.example.com", subdomainIndex),
			"host":       fmt.Sprintf("upstream-%d.internal:8080", subdomainIndex),
			"paths":      paths,
		})
	}

	payload["resources"] = resources

	raw, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("marshal parse benchmark service: %w", err))
	}

	return string(raw)
}
