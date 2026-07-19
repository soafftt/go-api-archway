package service

import (
	coreUpstream "core/domain/upstream"
	"core/gjwt"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gateway/controller/application/port/in/dto"
)

const lookupBenchmarkHS256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`

type benchmarkRouteCache struct {
	data map[string]*coreUpstream.UpstreamService
}

func (b *benchmarkRouteCache) LoadCache() {}

func (b *benchmarkRouteCache) Get(key string) (*coreUpstream.UpstreamService, bool) {
	service, ok := b.data[key]
	return service, ok
}

func (b *benchmarkRouteCache) Update(keys []string) error {
	return nil
}

func (b *benchmarkRouteCache) Evict(service string) {
	delete(b.data, service)
}

func BenchmarkUpstreamLookupServiceLookUpScaled(b *testing.B) {
	cases := []struct {
		name           string
		domains        int
		pathsPerDomain int
		withAuth       bool
	}{
		{
			name:           "4-domains-x-32-paths",
			domains:        4,
			pathsPerDomain: 32,
		},
		{
			name:           "16-domains-x-64-paths",
			domains:        16,
			pathsPerDomain: 64,
		},
		{
			name:           "16-domains-x-64-paths-with-hs256",
			domains:        16,
			pathsPerDomain: 64,
			withAuth:       true,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := fmt.Sprintf("lookup-bench-%d", time.Now().UnixNano())
			service, domain, requestPath, token := buildLookupBenchmarkService(b, serviceName, tc.domains, tc.pathsPerDomain, tc.withAuth)
			lookupService := NewUpstreamLookupService(&benchmarkRouteCache{
				data: map[string]*coreUpstream.UpstreamService{
					service.ServiceName: service,
				},
			})

			request := dto.UpStreamLookupRequest{
				Version:     "v1",
				Service:     service.ServiceName,
				Domain:      domain,
				Path:        requestPath,
				AccessToken: token,
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := lookupService.LookUpFromRequest(request)
				if result.Error != nil {
					b.Fatalf("lookup failed: %v", result.Error)
				}
				if result.Info.Host == "" || result.Info.Path == "" {
					b.Fatalf("lookup result info is incomplete: %+v", result.Info)
				}
			}
		})
	}
}

func buildLookupBenchmarkService(
	tb testing.TB,
	serviceName string,
	domains int,
	pathsPerDomain int,
	withAuth bool,
) (*coreUpstream.UpstreamService, string, string, *string) {
	tb.Helper()

	service := &coreUpstream.UpstreamService{
		ServiceName: serviceName,
		Resources:   make([]*coreUpstream.UpstreamResource, 0, domains),
	}

	if withAuth {
		service.Authorization = &struct {
			Algorithm string `json:"algorithm"`
			KeyData   string `json:"key_data"`
			UserKey   string `json:"user_key"`
		}{
			Algorithm: "HS256",
			KeyData:   base64.StdEncoding.EncodeToString([]byte(lookupBenchmarkHS256JWKJSON)),
			UserKey:   "user_id",
		}
		if err := gjwt.RegisterKeyByString(serviceName, service.Authorization.KeyData, gjwt.JSONKey, service.Authorization.Algorithm); err != nil {
			tb.Fatalf("register jwt key: %v", err)
		}
	}

	targetDomain := "api-0.example.com"
	targetPathSuffix := ((pathsPerDomain - 1) / 4) * 4
	for domainIndex := 0; domainIndex < domains; domainIndex++ {
		resource := &coreUpstream.UpstreamResource{
			Domain: fmt.Sprintf("api-%d.example.com", domainIndex),
			Host:   fmt.Sprintf("upstream-%d.internal:8080", domainIndex),
			Paths:  make([]*coreUpstream.UpstreamPath, 0, pathsPerDomain),
		}

		for pathIndex := 0; pathIndex < pathsPerDomain; pathIndex++ {
			path := &coreUpstream.UpstreamPath{
				Path:            fmt.Sprintf("/api/%d/items/%d", domainIndex, pathIndex),
				Method:          http.MethodGet,
				RequestTimeout:  int64(1000 + pathIndex),
				ResponseTimeout: int64(2000 + pathIndex),
				CacheTimeout:    int64(pathIndex % 10),
			}
			if pathIndex%4 == 0 {
				path.Path = fmt.Sprintf("/api/%d/users/{userId}/posts/%d", domainIndex, pathIndex)
			}
			if withAuth && domainIndex == 0 && pathIndex == targetPathSuffix {
				path.CheckAuthorization = true
			}
			resource.Paths = append(resource.Paths, path)
		}

		service.Resources = append(service.Resources, resource)
	}

	service.InitializeResourceIndex()

	requestPath := fmt.Sprintf("%s/api/0/users/123/posts/%d", targetDomain, targetPathSuffix)
	var token *string
	if withAuth {
		codec, err := gjwt.NewCodec(serviceName)
		if err != nil {
			tb.Fatalf("new jwt codec: %v", err)
		}
		now := time.Now()
		signed, err := codec.Serialize(nil, func(claims map[string]any) {
			claims["user_id"] = "bench-user"
			claims[string(gjwt.IssuedAt)] = now.Unix()
			claims[string(gjwt.Expiration)] = now.Add(time.Hour).Unix()
		})
		if err != nil {
			tb.Fatalf("sign jwt token: %v", err)
		}
		token = &signed
	}

	return service, targetDomain, requestPath, token
}
