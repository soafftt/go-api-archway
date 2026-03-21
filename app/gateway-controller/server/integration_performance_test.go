package server

import (
	"context"
	"encoding/json"
	upstreamDto "gateway/common/model/upstream"
	"gateway/controller/component"
	"gateway/controller/router"
	"gateway/controller/service"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// InMemoryRouteCache는 Valkey 없이 메모리만 사용하는 테스트용 캐시입니다
type InMemoryRouteCache struct {
	data map[string]*upstreamDto.UpstreamService
}

func NewInMemoryRouteCache() *InMemoryRouteCache {
	return &InMemoryRouteCache{
		data: make(map[string]*upstreamDto.UpstreamService),
	}
}

func (c *InMemoryRouteCache) Get(service string) (*upstreamDto.UpstreamService, bool) {
	svc, ok := c.data[service]
	return svc, ok
}

func (c *InMemoryRouteCache) Update(ctx context.Context, keys []string) error {
	return nil
}

func (c *InMemoryRouteCache) Evict(service string) {
	delete(c.data, service)
}

func (c *InMemoryRouteCache) LoadTestData() error {
	// 실제 프로덕션 데이터 구조
	testServices := []string{
		`{
			"service_name": "member-api",
			"resouces": [
				{
					"sub_domain": "user",
					"host": "http://user-service.internal:8080",
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
						},
						{
							"path": "/api/users/{userId}/posts",
							"method": "GET",
							"request_timeout": 5000,
							"response_timeout": 10000,
							"check_authorization": true
						},
						{
							"path": "/api/users/{userId}/posts/{postId}",
							"method": "GET",
							"request_timeout": 3000,
							"response_timeout": 5000,
							"check_authorization": true
						}
					]
				},
				{
					"sub_domain": "",
					"host": "http://default-service.internal:8081",
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
		}`,
		`{
			"service_name": "order-api",
			"resouces": [
				{
					"sub_domain": "orders",
					"host": "http://order-service.internal:9090",
					"paths": [
						{
							"path": "/api/orders",
							"method": "GET",
							"request_timeout": 5000,
							"response_timeout": 10000,
							"check_authorization": true
						},
						{
							"path": "/api/orders/{orderId}",
							"method": "GET",
							"request_timeout": 3000,
							"response_timeout": 5000,
							"check_authorization": true
						}
					]
				}
			]
		}`,
		`{
			"service_name": "payment-api",
			"resouces": [
				{
					"sub_domain": "payments",
					"host": "http://payment-service.internal:9091",
					"paths": [
						{
							"path": "/api/payments",
							"method": "POST",
							"request_timeout": 10000,
							"response_timeout": 15000,
							"check_authorization": true
						}
					]
				}
			]
		}`,
	}

	for _, jsonData := range testServices {
		var service upstreamDto.UpstreamService
		if err := json.Unmarshal([]byte(jsonData), &service); err != nil {
			return err
		}
		service.InitializeResourceIndex()
		c.data[service.ServiceName] = &service
	}

	return nil
}

// setupIntegrationTestServer는 실제 RouteService를 사용하는 통합 테스트 서버를 시작합니다
func setupIntegrationTestServer(socketPath string, routeCache component.RouteCache) (*http.Server, error) {
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	// 실제 RouteService 사용
	routeService := service.NewPolicyService(routeCache)
	controllerRouter := router.NewControllerRouter(routeService)

	server := &http.Server{
		Handler:      controllerRouter.Mux,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		IdleTimeout:  120 * time.Millisecond,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	return server, nil
}

// TestIntegrationRPSMeasurement는 실제 RouteService를 사용한 통합 성능 테스트입니다
func TestIntegrationRPSMeasurement(t *testing.T) {
	// 실제 메모리 캐시 설정
	cache := NewInMemoryRouteCache()
	if err := cache.LoadTestData(); err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	socketPath := "/tmp/gateway-controller-integration-rps.sock"
	server, err := setupIntegrationTestServer(socketPath, cache)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	defer func() {
		server.Shutdown(context.Background())
		os.Remove(socketPath)
	}()

	testCases := []struct {
		name        string
		duration    time.Duration
		concurrency int
		testURLs    []string
	}{
		{
			name:        "1 Client - 5s",
			duration:    5 * time.Second,
			concurrency: 1,
			testURLs: []string{
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/123",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders",
			},
		},
		{
			name:        "10 Clients - 5s",
			duration:    5 * time.Second,
			concurrency: 10,
			testURLs: []string{
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/123",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/payment-api/payments/api/payments",
			},
		},
		{
			name:        "50 Clients - 5s",
			duration:    5 * time.Second,
			concurrency: 50,
			testURLs: []string{
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/456",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/789/posts",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders/ord123",
			},
		},
		{
			name:        "100 Clients - 5s",
			duration:    5 * time.Second,
			concurrency: 100,
			testURLs: []string{
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/123",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/456/posts",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/789/posts/post123",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders",
				"http://unix/v1/uri/policy?upstream=http://example.com/v1/payment-api/payments/api/payments",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := createUnixSocketClient(socketPath)

			var (
				totalRequests  atomic.Int64
				successCount   atomic.Int64
				failureCount   atomic.Int64
				totalLatencyNs atomic.Int64
			)

			ctx, cancel := context.WithTimeout(context.Background(), tc.duration)
			defer cancel()

			var wg sync.WaitGroup
			startTime := time.Now()

			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					urlIndex := 0
					for {
						select {
						case <-ctx.Done():
							return
						default:
							// 여러 URL을 순회하며 테스트
							testURL := tc.testURLs[urlIndex%len(tc.testURLs)]
							urlIndex++

							reqStart := time.Now()
							resp, err := client.Get(testURL)
							latency := time.Since(reqStart)

							totalRequests.Add(1)
							totalLatencyNs.Add(latency.Nanoseconds())

							if err != nil {
								failureCount.Add(1)
								continue
							}

							io.Copy(io.Discard, resp.Body)
							resp.Body.Close()

							if resp.StatusCode == http.StatusOK {
								successCount.Add(1)
							} else {
								failureCount.Add(1)
								// 첫 10개의 실패만 로깅
								if failureCount.Load() <= 10 {
									t.Logf("Failed request: URL=%s, Status=%d", testURL, resp.StatusCode)
								}
							}
						}
					}
				}(i)
			}

			wg.Wait()
			elapsed := time.Since(startTime)

			total := totalRequests.Load()
			success := successCount.Load()
			failure := failureCount.Load()
			avgLatencyMs := float64(totalLatencyNs.Load()) / float64(total) / 1_000_000

			rps := float64(total) / elapsed.Seconds()

			t.Logf("\n========================================")
			t.Logf("🔥 INTEGRATION Performance Test: %s", tc.name)
			t.Logf("========================================")
			t.Logf("Duration:         %v", elapsed)
			t.Logf("Concurrency:      %d", tc.concurrency)
			t.Logf("Total Requests:   %d", total)
			t.Logf("Successful:       %d", success)
			t.Logf("Failed:           %d", failure)
			t.Logf("RPS:              %.2f requests/sec", rps)
			t.Logf("Avg Latency:      %.2f ms", avgLatencyMs)
			t.Logf("========================================\n")
		})
	}
}

// BenchmarkIntegrationUnixSocket는 실제 RouteService를 사용한 벤치마크입니다
func BenchmarkIntegrationUnixSocket(b *testing.B) {
	cache := NewInMemoryRouteCache()
	if err := cache.LoadTestData(); err != nil {
		b.Fatalf("Failed to load test data: %v", err)
	}

	socketPath := "/tmp/gateway-controller-integration-bench.sock"
	server, err := setupIntegrationTestServer(socketPath, cache)
	if err != nil {
		b.Fatalf("Failed to setup test server: %v", err)
	}
	defer func() {
		server.Shutdown(context.Background())
		os.Remove(socketPath)
	}()

	client := createUnixSocketClient(socketPath)
	testURLs := []string{
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/123",
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testURL := testURLs[i%len(testURLs)]
		resp, err := client.Get(testURL)
		if err != nil {
			b.Errorf("Request failed: %v", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkIntegrationParallel는 병렬 실행 벤치마크입니다
func BenchmarkIntegrationParallel(b *testing.B) {
	cache := NewInMemoryRouteCache()
	if err := cache.LoadTestData(); err != nil {
		b.Fatalf("Failed to load test data: %v", err)
	}

	socketPath := "/tmp/gateway-controller-integration-parallel.sock"
	server, err := setupIntegrationTestServer(socketPath, cache)
	if err != nil {
		b.Fatalf("Failed to setup test server: %v", err)
	}
	defer func() {
		server.Shutdown(context.Background())
		os.Remove(socketPath)
	}()

	client := createUnixSocketClient(socketPath)
	testURLs := []string{
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users",
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/123",
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/member-api/user/api/users/456/posts",
		"http://unix/v1/uri/policy?upstream=http://example.com/v1/order-api/orders/api/orders",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			testURL := testURLs[i%len(testURLs)]
			i++

			resp, err := client.Get(testURL)
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
