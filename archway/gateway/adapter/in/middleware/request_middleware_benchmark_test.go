package middleware

import (
	"context"
	"encoding/base64"
	"fmt"
	ingateway "gateway/adapter/in"
	"gateway/adapter/out/ratelimit"
	portIn "gateway/application/port/in"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"core/gjwt"
)

const (
	benchmarkHS256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`
	benchmarkRSAJWKBase64 = "eyJkIjoiS3V3SjFjbWRzc1pEMjkxTDVxdk5wMVZQTzMySHg4Q0hEekY1dzhyd1BYVjJvYWtVU3ZUckdhZnFzM0M1UkJkUUtZemZ1WGpfU0EwdGxjOEpObmhSQWhVd25pdG00QWg5dVZtRHlLZGtVZ1FCS0ZNVVQyWkRTdTJaT2daVEQ4OW1GcnNZX1JXSEJRMU1BcHlwbDd5ZDAyY3ZzV2dtbG9yeEVZTHNESGs0NGJ1cEkzMk9FZG05eWJjVmFlX2owTFktb3J1Q3IxNEVmdm1XWGFreXFOd204bXNkRkRYZkQ1TURpXy1MQURHNFVFa18wR29jeldlbEZQYUFCbHNhQ1BNRVUycnNEVEwzWnVUM1FCcmNCNzZaUVFEM1BFRlZCQS1FQTNtQnE4RlBrRVFqTF9IY3pUY1FYU2pfbGtsQ1pWd2ZiX21lQzNzaW5LU2w3TG5WVHo0b1F3IiwiZHAiOiJ0TXBUXzlJcVZDb2I2U1k4YUJrU3IyTWwtVXlPQk1TY29wSGRuWExsQ2ZDckVpdXo1YWdLUDVwTzhmNHktbW1nbHE0OEVpanZjY1cyUERHb3lvRzVnN280blIzM3ByOGFmcDlrQWx1QmZmWWpVZXA2WjBFdkVlaEo1RmtmUlA2UVFxbTV3M2ZVcGRPU19wb2RvTDE5bUkzTlVHRTltenpLVm9lamREel92YWMiLCJkcSI6IlpiSXJVdW5HZ2FEQW5rQXhLa0tObUpKdWp3LTJUc0Q5Sk0xUmxIdmZ1XzVxajdRS1dsdEJ4OTBxOWJHOWpXeDJnSHJCeWpPZVdWcXMxZEdUMVhOblVscUpaNWNHTjh4dXpXQmhBLXdjMUhwazhuNXpBNUg3RTlMbzdqa1puOXpaX1IyTEMwNUljS1ppTXBVaHBFUktsR0ZSUGZNb3RRU2cwV002b2laenFRayIsImUiOiJBUUFCIiwia3R5IjoiUlNBIiwibiI6IngxcU4wU0NfT3E1cVh6YjZCWDNuNk1JLWVxem5NZnFwcFVkcTBhWkpMUTB6U0RZSER0OGl4OUp1dGszQTVaOUpiZXJIT3JIdjlvQnNubmZRUnJZRHNLODdyN09hU0dVeHpyRm1ZLTZqRzg4dWZ3VWpBOEdfVl9HaVNYeTQ2VFpZRlpNb2J4UHVjdFJwNmhQeFBUNXQ5a19td2FIYnZ1Vm5zNmYyNnZOVmVrbko2YjhpLWlrbmFxR255VEhNSFNHUmtqX0FuVXlxbGF6cEFZdDZhSGZCQ2lWVllBUERzUUM3ZzYwQ01nNkNPX2hHVDFqRWRZY1c3VDZ6T1FKejE5cURIY0JkODE3dVVZQTEzR2tVaUoycEVJQUM5OTZWOWdKMFoyZGFNa0VVS0wxUlZTZk1mNm9uNURUMkswR0pibEpKVXludER1MUJDMTdBY1RxQ2FvbkdhUSIsInAiOiIzWHFaU0s5WUg0Y2FxajQ1NG9BMlA1X3Z3RTliUFFWRUNkbV9RRnd5T3d5a2tpMEtLbTBpS1NDdWVWbjdvUV92WXdrNWt2S09ZRWJJejMyaU12VGRXMEp0TlJFd0RfVTExQ1dXUGtyOE0tUGpPeDhoUjlTVVlQS0p0d0FGcUNRTWl4dnFMMGJua2I5VnptdDlKUlkxMDBfVk1RSlNJTHZEb0EzbjhzdHhXbGMiLCJxIjoiNW0wZkZfRTJ6SHRkUjBZT3Y5alh4ampycUFzWDRpOHhBNGJXRURFdFI0aUpmVFZlNXM2bU5idU9LenZBLVp3Z2l6ZGE3SGhPQUc4VU9EVEQ2WXcwQmRLVFhkdUQ3LUV2Mkh1ZERRblFOZ1N2a213SFBYZjRIdjdOajRIQVNHdmFzOWYtOGdjTEFEekc1ejlOZ29OVjI4Z2M0ZWt3UVlzQ1NKeTYxYUhPN1Q4IiwicWkiOiJneWVJbWpCMW5KRFE2ZWpmaGJLRGtReFJ1MjlKRC1RemJaSWdZVlkzOHYtN3pyMTl5TDRBSTVCOUMzSzF2VzJibVkwR1VENXRlNUdsQ2lDbHZ6NmpIVzBrRlNIUmY5V3VnSkVSVDQxQ081X25GZ1FMaUxnUXB4VGV3MjEyRVlfajB1aGJIMkhMRnp2blFNcUtNTlcyYmFUMlhFYy10dEFDaE1pb0Q1T0pXalEifQ=="
)

type benchmarkScenario struct {
	name             string
	uriCount         int
	rateLimitPattern string
}

type benchmarkRoute struct {
	targetPath    string
	authRequired  bool
	rateLimit     int64
	lookupResult  portIn.UpstreamLookupResult
	requestSample *http.Request
}

type benchmarkFixture struct {
	routes []benchmarkRoute
	lookup map[string]benchmarkRoute
}

type benchmarkLookupUseCase struct {
	lookupTable map[string]benchmarkRoute
	jwtCodec    gjwt.Codec
}

func (b benchmarkLookupUseCase) Lookup(srcPath string, accessToken *string) (portIn.UpstreamLookupResult, error) {
	route, ok := b.lookupTable[srcPath]
	if !ok {
		return portIn.UpstreamLookupResult{}, fmt.Errorf("route not found: %s", srcPath)
	}

	if route.authRequired {
		if accessToken == nil {
			return portIn.UpstreamLookupResult{}, fmt.Errorf("access token required")
		}
		result := b.jwtCodec.Parse(*accessToken)
		if result.Err != nil || !result.Valid {
			return portIn.UpstreamLookupResult{}, fmt.Errorf("invalid jwt token")
		}
	}

	return route.lookupResult, nil
}

func BenchmarkRequestMiddleware_RateLimitMatrix(b *testing.B) {
	scenarios := []benchmarkScenario{
		{name: "URI100_AllRateLimitDistinct", uriCount: 100, rateLimitPattern: "all"},
		{name: "URI100_RateLimit_2_to_3", uriCount: 100, rateLimitPattern: "2:3"},
		{name: "URI100_RateLimit_1_to_3", uriCount: 100, rateLimitPattern: "1:3"},
		{name: "URI1000_AllRateLimitDistinct", uriCount: 1000, rateLimitPattern: "all"},
		{name: "URI1000_RateLimit_2_to_3", uriCount: 1000, rateLimitPattern: "2:3"},
		{name: "URI1000_RateLimit_1_to_3", uriCount: 1000, rateLimitPattern: "1:3"},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		parallelLevels := []int{1, 8, 32}
		if scenario.uriCount == 1000 {
			parallelLevels = []int{1, 16, 64}
		}

		for _, algorithm := range []string{"HS256", "RS256"} {
			algorithm := algorithm
			b.Run(fmt.Sprintf("%s_%s", scenario.name, algorithm), func(b *testing.B) {
				fixture, token := buildBenchmarkFixture(b, scenario, algorithm, "")
				lookupUseCase := benchmarkLookupUseCase{
					lookupTable: fixture.lookup,
					jwtCodec:    mustNewCodec(b, buildJWTKeyName(scenario.name, algorithm)),
				}

				for _, parallelLevel := range parallelLevels {
					parallelLevel := parallelLevel
					b.Run(fmt.Sprintf("P%d", parallelLevel), func(b *testing.B) {
						middlewareHandler := NewRequestMiddleware(lookupUseCase, ratelimit.NewRateLimit()).HandleMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
							writer.WriteHeader(http.StatusOK)
						}))
						runMiddlewareBenchmark(b, middlewareHandler, fixture.routes, token, parallelLevel, false)
					})
				}
			})
		}
	}
}

func BenchmarkGatewayPipeline_RateLimitMatrixE2E(b *testing.B) {
	scenarios := []benchmarkScenario{
		{name: "URI100_AllRateLimitDistinct", uriCount: 100, rateLimitPattern: "all"},
		{name: "URI100_RateLimit_2_to_3", uriCount: 100, rateLimitPattern: "2:3"},
		{name: "URI100_RateLimit_1_to_3", uriCount: 100, rateLimitPattern: "1:3"},
		{name: "URI1000_AllRateLimitDistinct", uriCount: 1000, rateLimitPattern: "all"},
		{name: "URI1000_RateLimit_2_to_3", uriCount: 1000, rateLimitPattern: "2:3"},
		{name: "URI1000_RateLimit_1_to_3", uriCount: 1000, rateLimitPattern: "1:3"},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		parallelLevel := 4
		if scenario.uriCount == 1000 {
			parallelLevel = 8
		}

		for _, algorithm := range []string{"HS256", "RS256"} {
			algorithm := algorithm
			b.Run(fmt.Sprintf("%s_%s", scenario.name, algorithm), func(b *testing.B) {
				upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte("ok"))
				}))
				defer upstream.Close()

				fixture, token := buildBenchmarkFixture(b, scenario, algorithm, upstream.URL)
				lookupUseCase := benchmarkLookupUseCase{
					lookupTable: fixture.lookup,
					jwtCodec:    mustNewCodec(b, buildJWTKeyName(scenario.name, algorithm)),
				}

				proxy := ingateway.NewGatewayProxy()
				handler := NewRequestMiddleware(lookupUseCase, ratelimit.NewRateLimit()).HandleMiddleware(proxy.HttpProxy)
				runMiddlewareBenchmark(b, handler, fixture.routes, token, parallelLevel, true)
			})
		}
	}
}

func runMiddlewareBenchmark(b *testing.B, handler http.Handler, routes []benchmarkRoute, jwtToken string, parallelLevel int, failOnServerError bool) {
	b.Helper()
	b.ReportAllocs()
	b.SetParallelism(parallelLevel)

	const maxWindows = 300
	var requestCountBySecond [maxWindows]atomic.Int64
	var cursor atomic.Uint64
	var serverErrors atomic.Uint64
	startedAt := time.Now()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		responseWriter := &benchmarkResponseWriter{}
		for pb.Next() {
			index := int(cursor.Add(1)-1) % len(routes)
			route := routes[index]

			request := route.requestSample.Clone(context.Background())
			if route.authRequired {
				request.Header.Set("Authorization", "Bearer "+jwtToken)
			}

			responseWriter.reset()
			handler.ServeHTTP(responseWriter, request)
			if failOnServerError && responseWriter.statusCode >= http.StatusInternalServerError {
				serverErrors.Add(1)
			}

			if second := int(time.Since(startedAt) / time.Second); second >= 0 && second < maxWindows {
				requestCountBySecond[second].Add(1)
			}
		}
	})
	b.StopTimer()
	if failOnServerError && serverErrors.Load() > 0 {
		b.Fatalf("observed %d 5xx responses during benchmark", serverErrors.Load())
	}

	throughputSamples := collectRpsSamples(requestCountBySecond[:])
	if len(throughputSamples) == 0 {
		return
	}
	b.ReportMetric(percentile(throughputSamples, 0.50), "rps_p50")
	b.ReportMetric(percentile(throughputSamples, 0.95), "rps_p95")
	b.ReportMetric(percentile(throughputSamples, 0.99), "rps_p99")
}

func buildBenchmarkFixture(tb testing.TB, scenario benchmarkScenario, algorithm string, upstreamHost string) (benchmarkFixture, string) {
	tb.Helper()

	keyName := buildJWTKeyName(scenario.name, algorithm)
	registerBenchmarkKey(tb, keyName, algorithm)

	tokenCodec := mustNewCodec(tb, keyName)
	token := mustSignBenchmarkToken(tb, tokenCodec)

	activeRateLimitCount := scenario.uriCount
	switch scenario.rateLimitPattern {
	case "2:3":
		activeRateLimitCount = scenario.uriCount * 2 / 5
	case "1:3":
		activeRateLimitCount = scenario.uriCount / 4
	}

	routes := make([]benchmarkRoute, 0, scenario.uriCount)
	lookup := make(map[string]benchmarkRoute, scenario.uriCount)
	for index := 0; index < scenario.uriCount; index++ {
		targetPath := fmt.Sprintf("/bench/path-%04d", index)
		originPath := fmt.Sprintf("/origin/path-%04d", index)
		upstreamPath := fmt.Sprintf("/upstream/path-%04d", index)
		if upstreamHost == "" {
			upstreamPath = "/blackhole"
		}

		rateLimitCount := int64(0)
		if isRateLimitActive(index, scenario.uriCount, activeRateLimitCount) {
			rateLimitCount = int64(1000 + index)
		}

		route := benchmarkRoute{
			targetPath:   targetPath,
			authRequired: index%2 == 0,
			rateLimit:    rateLimitCount,
			lookupResult: portIn.UpstreamLookupResult{
				ServiceName:    "svc-" + scenario.name + "-" + algorithm,
				Host:           upstreamHost,
				Path:           upstreamPath,
				OriginPath:     originPath,
				Method:         http.MethodGet,
				RateLimitCount: rateLimitCount,
			},
			requestSample: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: targetPath},
				Header: make(http.Header, 1),
				Host:   "gateway.local",
			},
		}

		routes = append(routes, route)
		lookup[targetPath] = route
	}

	return benchmarkFixture{
		routes: routes,
		lookup: lookup,
	}, token
}

func isRateLimitActive(index int, total int, activeCount int) bool {
	if activeCount <= 0 {
		return false
	}
	if activeCount >= total {
		return true
	}
	// deterministic shuffle without extra allocations
	shuffled := (index * 2654435761) % total
	return shuffled < activeCount
}

func buildJWTKeyName(caseName string, algorithm string) string {
	return "gateway-bench-" + caseName + "-" + algorithm
}

func registerBenchmarkKey(tb testing.TB, keyName string, algorithm string) {
	tb.Helper()

	switch algorithm {
	case "HS256":
		if err := gjwt.RegisterKey(keyName, []byte(benchmarkHS256JWKJSON), gjwt.JSONKey, algorithm); err != nil {
			tb.Fatalf("register HS256 key: %v", err)
		}
	case "RS256":
		rsaJWK, err := base64.StdEncoding.DecodeString(benchmarkRSAJWKBase64)
		if err != nil {
			tb.Fatalf("decode RS256 key: %v", err)
		}
		if err := gjwt.RegisterKey(keyName, rsaJWK, gjwt.JSONKey, algorithm); err != nil {
			tb.Fatalf("register RS256 key: %v", err)
		}
	default:
		tb.Fatalf("unsupported algorithm: %s", algorithm)
	}
}

func mustNewCodec(tb testing.TB, keyName string) gjwt.Codec {
	tb.Helper()
	codec, err := gjwt.NewCodec(keyName)
	if err != nil {
		tb.Fatalf("new jwt codec: %v", err)
	}
	return codec
}

func mustSignBenchmarkToken(tb testing.TB, codec gjwt.Codec) string {
	tb.Helper()
	now := time.Now()
	token, err := codec.Serialize(nil, func(claims map[string]any) {
		claims["user_id"] = "benchmark-user"
		claims[string(gjwt.IssuedAt)] = now.Unix()
		claims[string(gjwt.Expiration)] = now.Add(1 * time.Hour).Unix()
	})
	if err != nil {
		tb.Fatalf("sign jwt token: %v", err)
	}
	return token
}

func collectRpsSamples(windows []atomic.Int64) []float64 {
	samples := make([]float64, 0, len(windows))
	for _, window := range windows {
		value := window.Load()
		if value > 0 {
			samples = append(samples, float64(value))
		}
	}
	return samples
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 1 {
		return values[len(values)-1]
	}

	sorted := slices.Clone(values)
	slices.Sort(sorted)
	index := int(math.Ceil(p*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

type benchmarkResponseWriter struct {
	headers    http.Header
	statusCode int
}

func (b *benchmarkResponseWriter) Header() http.Header {
	if b.headers == nil {
		b.headers = make(http.Header)
	}
	return b.headers
}

func (b *benchmarkResponseWriter) Write(data []byte) (int, error) {
	if b.statusCode == 0 {
		b.statusCode = http.StatusOK
	}
	return len(data), nil
}

func (b *benchmarkResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

func (b *benchmarkResponseWriter) reset() {
	b.statusCode = 0
	clear(b.headers)
}
