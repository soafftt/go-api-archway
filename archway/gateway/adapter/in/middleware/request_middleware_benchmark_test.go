package middleware

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ingateway "gateway/adapter/in"
	"gateway/adapter/out/ratelimit"
	portIn "gateway/application/port/in"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"core/gjwt"
	pb "protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

func (b benchmarkLookupUseCase) Lookup(srcPath string, accessToken *string, _ portIn.Transport) (portIn.UpstreamLookupResult, error) {
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

type benchmarkGatewayControllerGrpcLookupUseCase struct {
	serviceClient pb.UpstreamLookupServiceClient
}

type benchmarkGatewayControllerUnixHTTPLookupUseCase struct {
	baseURL string
	client  *http.Client
}

func (u benchmarkGatewayControllerGrpcLookupUseCase) Lookup(srcPath string, accessToken *string, _ portIn.Transport) (portIn.UpstreamLookupResult, error) {
	var result portIn.UpstreamLookupResult

	requestContext := context.Background()
	if accessToken != nil {
		requestContext = metadata.NewOutgoingContext(requestContext, metadata.Pairs("authorization", *accessToken))
	}

	response, err := u.serviceClient.Lookup(requestContext, &pb.UpstreamLookupRequest{Path: srcPath})
	if err != nil {
		return result, err
	}
	if response.GetError() != "" {
		return result, fmt.Errorf("gateway-controller business error: %s", response.GetError())
	}
	info := response.GetInfo()
	if info == nil {
		return result, fmt.Errorf("gateway-controller returned empty info")
	}

	result = portIn.UpstreamLookupResult{
		ServiceName:     info.GetServiceName(),
		Host:            info.GetHost(),
		Path:            info.GetPath(),
		OriginPath:      info.GetOriginalPath(),
		Method:          info.GetMethod(),
		ResponseTimeout: info.GetResponseTimeout(),
		RequestTimeout:  info.GetRequestTimeout(),
		CacheTimeout:    info.GetCacheTimeout(),
		UserKey:         info.GetUserKey(),
		RateLimitCount:  info.GetRateLimitCount(),
	}

	return result, nil
}

func (u benchmarkGatewayControllerUnixHTTPLookupUseCase) Lookup(srcPath string, accessToken *string, _ portIn.Transport) (portIn.UpstreamLookupResult, error) {
	var result portIn.UpstreamLookupResult

	request, err := http.NewRequest(http.MethodGet, u.baseURL+srcPath, nil)
	if err != nil {
		return result, err
	}
	if accessToken != nil {
		request.Header.Set("Authorization", "Bearer "+*accessToken)
	}

	response, err := u.client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("gateway-controller http status: %d", response.StatusCode)
	}

	decoded := struct {
		ServiceName     string `json:"service_name"`
		Host            string `json:"domain"`
		Path            string `json:"path"`
		OriginalPath    string `json:"original_path"`
		Method          string `json:"method"`
		ResponseTimeout int64  `json:"response_timeout"`
		RequestTimeout  int64  `json:"request_timeout"`
		CacheTimeout    int64  `json:"cache_timeout"`
		UserKey         any    `json:"user_key"`
		RateLimitCount  int64  `json:"rate_limit_count"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return result, err
	}

	result = portIn.UpstreamLookupResult{
		ServiceName:     decoded.ServiceName,
		Host:            decoded.Host,
		Path:            decoded.Path,
		OriginPath:      decoded.OriginalPath,
		Method:          decoded.Method,
		ResponseTimeout: decoded.ResponseTimeout,
		RequestTimeout:  decoded.RequestTimeout,
		CacheTimeout:    decoded.CacheTimeout,
		UserKey:         decoded.UserKey,
		RateLimitCount:  decoded.RateLimitCount,
	}
	return result, nil
}

func BenchmarkGatewayFullChainE2E_RewriteViaGatewayController(b *testing.B) {
	controllerRootDir := benchmarkResolveGatewayControllerRootDir(b)
	valkeyAddress := benchmarkResolveValkeyAddress()

	scenarios := []struct {
		name      string
		algorithm string
		withAuth  bool
	}{
		{name: "NoJWT", algorithm: "HS256", withAuth: false},
		{name: "HS256", algorithm: "HS256", withAuth: true},
		{name: "RS256", algorithm: "RS256", withAuth: true},
	}

	protocols := []struct {
		name            string
		serverTransport string
	}{
		{name: "gRPC", serverTransport: "grpc"},
		{name: "unixHTTP", serverTransport: "http"},
	}

	cases := []struct {
		name        string
		parallelism int
	}{
		{name: "P1", parallelism: 1},
		{name: "P8", parallelism: 8},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			for _, protocol := range protocols {
				protocol := protocol
				b.Run(protocol.name, func(b *testing.B) {
					for _, tc := range cases {
						tc := tc
						b.Run(tc.name, func(b *testing.B) {
							serviceName := "gateway-fullchain-bench-" + scenario.name + "-" + protocol.name + "-" + tc.name + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
							valkeyKey := "UPSTREAM:" + serviceName
							socketPath := filepath.Join(os.TempDir(), "gateway-controller-fullchain-"+fmt.Sprintf("%d", time.Now().UnixNano())+".sock")

							upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
								writer.Header().Set("X-Rewrite-Path", request.URL.Path)
								writer.WriteHeader(http.StatusOK)
								_, _ = writer.Write([]byte("ok"))
							}))
							defer upstream.Close()

							payload := benchmarkBuildGatewayControllerValkeyPayload(serviceName, upstream.URL, scenario.algorithm, scenario.withAuth)
							if err := benchmarkValkeySet(valkeyAddress, valkeyKey, payload); err != nil {
								b.Fatalf("failed to set valkey fixture: %v", err)
							}
							b.Cleanup(func() {
								_ = benchmarkValkeyDel(valkeyAddress, valkeyKey)
							})

							cmd := exec.Command("go", "run", "./cmd")
							cmd.Dir = controllerRootDir
							cmd.Env = append(os.Environ(),
								"VALKEY_MASTER_HOST="+valkeyAddress,
								"UNIX_SOCKET_PATH="+socketPath,
								"SERVER_TRANSPORT="+protocol.serverTransport,
							)
							if err := cmd.Start(); err != nil {
								b.Fatalf("failed to start gateway-controller: %v", err)
							}
							b.Cleanup(func() {
								_ = cmd.Process.Signal(os.Interrupt)
								done := make(chan error, 1)
								go func() { done <- cmd.Wait() }()
								select {
								case <-time.After(3 * time.Second):
									_ = cmd.Process.Kill()
								case <-done:
								}
								_ = os.Remove(socketPath)
							})

							benchmarkWaitForUnixSocketReady(b, socketPath)

							var lookupUseCase portIn.UpstreamLookupUseCase
							if protocol.serverTransport == "grpc" {
								connection, err := grpc.NewClient(
									"passthrough:///gateway-controller-bench",
									grpc.WithTransportCredentials(insecure.NewCredentials()),
									grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
										return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
									}),
								)
								if err != nil {
									b.Fatalf("failed to create grpc client: %v", err)
								}
								connection.Connect()
								b.Cleanup(func() {
									_ = connection.Close()
								})
								lookupUseCase = benchmarkGatewayControllerGrpcLookupUseCase{
									serviceClient: pb.NewUpstreamLookupServiceClient(connection),
								}
							} else {
								httpClient := benchmarkNewUnixSocketHTTPClient(socketPath)
								lookupUseCase = benchmarkGatewayControllerUnixHTTPLookupUseCase{
									baseURL: "http://unix/v1/upstream?path=",
									client:  httpClient,
								}
							}

							requestMiddleware := NewRequestMiddleware(lookupUseCase, ratelimit.NewRateLimit())
							handler := Chain(ingateway.NewGatewayProxy().HttpProxy, requestMiddleware)

							targetPath := "/v1/" + serviceName + "/api.example.com/echo/user-123/posts/777?trace=true"
							requestSample := httptest.NewRequest(http.MethodGet, "http://gateway.local"+targetPath, nil)

							token := ""
							if scenario.withAuth {
								keyName := "gateway-fullchain-token-" + serviceName
								registerBenchmarkKey(b, keyName, scenario.algorithm)
								token = mustSignBenchmarkToken(b, mustNewCodec(b, keyName))
							}

							routes := []benchmarkRoute{
								{
									targetPath:    targetPath,
									authRequired:  scenario.withAuth,
									rateLimit:     0,
									requestSample: requestSample,
								},
							}

							benchmarkWaitForGatewayFullChainReady(b, handler, requestSample, token, scenario.withAuth)
							runMiddlewareBenchmark(b, handler, routes, token, tc.parallelism, true)
						})
					}
				})
			}
		})
	}
}

func BenchmarkGatewayFullChainRateLimitMatrixE2E(b *testing.B) {
	controllerRootDir := benchmarkResolveGatewayControllerRootDir(b)
	valkeyAddress := benchmarkResolveValkeyAddress()

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
				serviceName := "gateway-fullchain-matrix-" + scenario.name + "-" + algorithm + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
				valkeyKey := "UPSTREAM:" + serviceName
				socketPath := filepath.Join(os.TempDir(), "gateway-controller-matrix-"+fmt.Sprintf("%d", time.Now().UnixNano())+".sock")

				upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(http.StatusOK)
					_, _ = writer.Write([]byte("ok"))
				}))
				defer upstream.Close()

				routes, token, payload := buildFullChainMatrixFixture(b, scenario, algorithm, serviceName, upstream.URL)
				if err := benchmarkValkeySet(valkeyAddress, valkeyKey, payload); err != nil {
					b.Fatalf("failed to set valkey fixture: %v", err)
				}
				b.Cleanup(func() {
					_ = benchmarkValkeyDel(valkeyAddress, valkeyKey)
				})

				cmd := exec.Command("go", "run", "./cmd")
				cmd.Dir = controllerRootDir
				cmd.Env = append(os.Environ(),
					"VALKEY_MASTER_HOST="+valkeyAddress,
					"UNIX_SOCKET_PATH="+socketPath,
					"SERVER_TRANSPORT=grpc",
				)
				if err := cmd.Start(); err != nil {
					b.Fatalf("failed to start gateway-controller: %v", err)
				}
				b.Cleanup(func() {
					_ = cmd.Process.Signal(os.Interrupt)
					done := make(chan error, 1)
					go func() { done <- cmd.Wait() }()
					select {
					case <-time.After(3 * time.Second):
						_ = cmd.Process.Kill()
					case <-done:
					}
					_ = os.Remove(socketPath)
				})

				benchmarkWaitForUnixSocketReady(b, socketPath)

				connection, err := grpc.NewClient(
					"passthrough:///gateway-controller-matrix",
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
						return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
					}),
				)
				if err != nil {
					b.Fatalf("failed to create grpc client: %v", err)
				}
				connection.Connect()
				b.Cleanup(func() {
					_ = connection.Close()
				})

				lookupUseCase := benchmarkGatewayControllerGrpcLookupUseCase{
					serviceClient: pb.NewUpstreamLookupServiceClient(connection),
				}

				requestMiddleware := NewRequestMiddleware(lookupUseCase, ratelimit.NewRateLimit())
				handler := Chain(ingateway.NewGatewayProxy().HttpProxy, requestMiddleware)

				benchmarkWaitForGatewayFullChainReady(b, handler, routes[0].requestSample, token, routes[0].authRequired)
				runMiddlewareBenchmark(b, handler, routes, token, parallelLevel, true)
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

func benchmarkWaitForGatewayFullChainReady(tb testing.TB, handler http.Handler, requestSample *http.Request, token string, withAuth bool) {
	tb.Helper()

	deadline := time.Now().Add(4 * time.Second)
	lastStatusCode := 0
	lastBody := ""
	for time.Now().Before(deadline) {
		request := requestSample.Clone(context.Background())
		if withAuth {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		lastStatusCode = response.Code
		lastBody = response.Body.String()
		if response.Code == http.StatusOK {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	tb.Fatalf("gateway full-chain benchmark target not ready: path=%s status=%d body=%q", requestSample.URL.Path, lastStatusCode, lastBody)
}

func benchmarkResolveGatewayControllerRootDir(tb testing.TB) string {
	tb.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("failed to resolve benchmark file path")
	}

	gatewayRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	controllerRoot := filepath.Join(filepath.Dir(gatewayRoot), "gateway-controller")
	if info, err := os.Stat(filepath.Join(controllerRoot, "cmd")); err != nil || !info.IsDir() {
		tb.Fatalf("gateway-controller root not found: %s", controllerRoot)
	}

	return controllerRoot
}

func benchmarkResolveValkeyAddress() string {
	host := strings.TrimSpace(os.Getenv("VALKEY_MASTER_HOST"))
	if host == "" {
		return "127.0.0.1:6379"
	}
	return host
}

func benchmarkBuildGatewayControllerValkeyPayload(serviceName, upstreamHost, algorithm string, withAuth bool) string {
	authorizationBlock := ""
	if withAuth {
		authorizationBlock = fmt.Sprintf(`
	"authorization": {
		"algorithm": %q,
		"key_data": %q,
		"user_key": "user_id"
	},`, algorithm, benchmarkJWTKeyDataForAlgorithm(algorithm))
	}

	return fmt.Sprintf(`{
	"service_name": %q,
	%s
	"resources": [
		{
			"domain": "api.example.com",
			"host": %q,
			"paths": [
				{
					"path": "/echo/{userId}/posts/{postId}",
					"method": "GET",
					"request_timeout": 3000,
					"response_timeout": 5000,
					"check_authorization": %t,
					"cache_timeout": 0,
					"rate_limit_count": 0
				}
			]
		}
	]
}`, serviceName, authorizationBlock, upstreamHost, withAuth)
}

func benchmarkValkeySet(addr, key, value string) error {
	connection, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return err
	}
	defer connection.Close()

	command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
	if _, err := connection.Write([]byte(command)); err != nil {
		return err
	}

	reader := bufio.NewReader(connection)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "+OK") {
		return fmt.Errorf("unexpected valkey SET response: %s", strings.TrimSpace(line))
	}

	return nil
}

func benchmarkValkeyDel(addr, key string) error {
	connection, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return err
	}
	defer connection.Close()

	command := fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
	if _, err := connection.Write([]byte(command)); err != nil {
		return err
	}

	reader := bufio.NewReader(connection)
	_, err = reader.ReadString('\n')
	return err
}

func benchmarkWaitForUnixSocketReady(tb testing.TB, socketPath string) {
	tb.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	tb.Fatalf("gateway-controller unix socket not ready: %s", socketPath)
}

func benchmarkNewUnixSocketHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}
}

func benchmarkJWTKeyDataForAlgorithm(algorithm string) string {
	switch algorithm {
	case "HS256":
		return base64.StdEncoding.EncodeToString([]byte(benchmarkHS256JWKJSON))
	case "RS256":
		return benchmarkRSAJWKBase64
	default:
		panic("unsupported algorithm: " + algorithm)
	}
}

func buildFullChainMatrixFixture(tb testing.TB, scenario benchmarkScenario, algorithm, serviceName, upstreamHost string) ([]benchmarkRoute, string, string) {
	tb.Helper()

	keyName := "gateway-fullchain-matrix-key-" + serviceName
	registerBenchmarkKey(tb, keyName, algorithm)
	token := mustSignBenchmarkToken(tb, mustNewCodec(tb, keyName))

	activeRateLimitCount := scenario.uriCount
	switch scenario.rateLimitPattern {
	case "2:3":
		activeRateLimitCount = scenario.uriCount * 2 / 5
	case "1:3":
		activeRateLimitCount = scenario.uriCount / 4
	}

	type payloadAuthorization struct {
		Algorithm string `json:"algorithm"`
		KeyData   string `json:"key_data"`
		UserKey   string `json:"user_key"`
	}
	type payloadPath struct {
		Path               string `json:"path"`
		Method             string `json:"method"`
		RequestTimeout     int64  `json:"request_timeout"`
		ResponseTimeout    int64  `json:"response_timeout"`
		CheckAuthorization bool   `json:"check_authorization"`
		CacheTimeout       int64  `json:"cache_timeout"`
		RateLimitCount     int64  `json:"rate_limit_count"`
	}
	type payloadResource struct {
		Domain string        `json:"domain"`
		Host   string        `json:"host"`
		Paths  []payloadPath `json:"paths"`
	}
	type payloadService struct {
		ServiceName   string               `json:"service_name"`
		Authorization payloadAuthorization `json:"authorization"`
		Resources     []payloadResource    `json:"resources"`
	}

	routes := make([]benchmarkRoute, 0, scenario.uriCount)
	paths := make([]payloadPath, 0, scenario.uriCount)
	for index := 0; index < scenario.uriCount; index++ {
		targetPath := fmt.Sprintf("/bench/path-%04d", index)
		fullPath := "/v1/" + serviceName + "/api.example.com" + targetPath
		authRequired := index%2 == 0
		rateLimitCount := int64(0)
		if isRateLimitActive(index, scenario.uriCount, activeRateLimitCount) {
			rateLimitCount = int64(1000 + index)
		}

		routes = append(routes, benchmarkRoute{
			targetPath:   fullPath,
			authRequired: authRequired,
			rateLimit:    rateLimitCount,
			requestSample: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: fullPath},
				Header: make(http.Header, 1),
				Host:   "gateway.local",
			},
		})

		paths = append(paths, payloadPath{
			Path:               targetPath,
			Method:             http.MethodGet,
			RequestTimeout:     3000,
			ResponseTimeout:    5000,
			CheckAuthorization: authRequired,
			CacheTimeout:       0,
			RateLimitCount:     rateLimitCount,
		})
	}

	payload := payloadService{
		ServiceName: serviceName,
		Authorization: payloadAuthorization{
			Algorithm: algorithm,
			KeyData:   benchmarkJWTKeyDataForAlgorithm(algorithm),
			UserKey:   "user_id",
		},
		Resources: []payloadResource{
			{
				Domain: "api.example.com",
				Host:   upstreamHost,
				Paths:  paths,
			},
		},
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		tb.Fatalf("failed to marshal matrix payload: %v", err)
	}

	return routes, token, string(encoded)
}
