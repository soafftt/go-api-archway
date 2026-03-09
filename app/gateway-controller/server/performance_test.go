package server

import (
	"context"
	"encoding/json"
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

type MockRouteService struct{}

func (m *MockRouteService) GetRouteInfo(urlParseDto interface{}) (interface{}, error) {
	return map[string]string{
		"service": "test-service",
		"path":    "/test/path",
		"version": "v1",
	}, nil
}

func setupTestServer(socketPath string) (*http.Server, error) {
	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	mockService := &MockRouteService{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/uri/policy", func(w http.ResponseWriter, r *http.Request) {
		targetUrl := r.URL.Query().Get("upstream")
		if targetUrl == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		info, _ := mockService.GetRouteInfo(nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)
	})
	server := &http.Server{
		Handler:      mux,
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

func createUnixSocketClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
}

func TestRPSMeasurement(t *testing.T) {
	socketPath := "/tmp/gateway-controller-rps.sock"
	server, err := setupTestServer(socketPath)
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
	}{
		{"1 Client - 5s", 5 * time.Second, 1},
		{"10 Clients - 5s", 5 * time.Second, 10},
		{"50 Clients - 5s", 5 * time.Second, 50},
		{"100 Clients - 5s", 5 * time.Second, 100},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := createUnixSocketClient(socketPath)
			testURL := "http://unix/v1/uri/policy?upstream=http://example.com/v1/users/profile/123"
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
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							return
						default:
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
							}
						}
					}
				}()
			}
			wg.Wait()
			elapsed := time.Since(startTime)
			total := totalRequests.Load()
			success := successCount.Load()
			failure := failureCount.Load()
			avgLatencyMs := float64(totalLatencyNs.Load()) / float64(total) / 1_000_000
			rps := float64(total) / elapsed.Seconds()
			t.Logf("========================================")
			t.Logf("Performance Test Results: %s", tc.name)
			t.Logf("========================================")
			t.Logf("Duration:         %v", elapsed)
			t.Logf("Concurrency:      %d", tc.concurrency)
			t.Logf("Total Requests:   %d", total)
			t.Logf("Successful:       %d", success)
			t.Logf("Failed:           %d", failure)
			t.Logf("RPS:              %.2f requests/sec", rps)
			t.Logf("Avg Latency:      %.2f ms", avgLatencyMs)
			t.Logf("========================================")
		})
	}
}

func BenchmarkUnixSocketSingleRequest(b *testing.B) {
	socketPath := "/tmp/gateway-controller-bench.sock"
	server, err := setupTestServer(socketPath)
	if err != nil {
		b.Fatalf("Failed to setup test server: %v", err)
	}
	defer func() {
		server.Shutdown(context.Background())
		os.Remove(socketPath)
	}()

	client := createUnixSocketClient(socketPath)
	testURL := "http://unix/v1/uri/policy?upstream=http://example.com/v1/users/profile/123"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(testURL)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkUnixSocketParallelRequests(b *testing.B) {
	socketPath := "/tmp/gateway-controller-bench-parallel.sock"
	server, err := setupTestServer(socketPath)
	if err != nil {
		b.Fatalf("Failed to setup test server: %v", err)
	}
	defer func() {
		server.Shutdown(context.Background())
		os.Remove(socketPath)
	}()

	client := createUnixSocketClient(socketPath)
	testURL := "http://unix/v1/uri/policy?upstream=http://example.com/v1/users/profile/123"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
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