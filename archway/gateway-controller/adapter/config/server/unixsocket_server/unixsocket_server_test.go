package unixsocket_server

import (
	"context"
	coreAdapterIn "core/adapter/in"
	coreUpstream "core/domain/upstream"
	"core/gjwt"
	coreUtils "core/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	app_config "gateway/controller/adapter/config/app_config"
	valkey2 "gateway/controller/adapter/config/valkey"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	adapterPortInUnix "gateway/controller/adapter/port/in/unix"
	adapterPortInUnixDi "gateway/controller/adapter/port/in/unix/di"
	adapterPortInUnixHandler "gateway/controller/adapter/port/in/unix/handler"
	applicationPortIn "gateway/controller/application/port/in"
	applicationDto "gateway/controller/application/port/in/dto"
	applicationCache "gateway/controller/application/port/out/cache"
	applicationService "gateway/controller/application/service"

	"github.com/joho/godotenv"
	"github.com/valkey-io/valkey-go"
)

const unixServerBenchmarkServiceJSON = `{
	"service_name": "member-api",
	"resources": [
		{
			"domain": "api.example.com",
			"host": "upstream-server-1.internal:8080",
			"paths": [
				{
					"path": "/api/users",
					"method": "GET",
					"request_timeout": 5000,
					"response_timeout": 10000,
					"check_authorization": false,
					"cache_timeout": 10
				},
				{
					"path": "/api/users/{userId}/posts/{postId}",
					"method": "GET",
					"request_timeout": 3000,
					"response_timeout": 5000,
					"check_authorization": false
				}
			]
		},
		{
			"domain": "",
			"host": "default-user.internal:8081",
			"paths": [
				{
					"path": "/v1/member/",
					"method": "GET",
					"request_timeout": 1500,
					"response_timeout": 2500,
					"check_authorization": false
				}
			]
		}
	]
}`

const rsaJWKBase64 = "eyJkIjoiS3V3SjFjbWRzc1pEMjkxTDVxdk5wMVZQTzMySHg4Q0hEekY1dzhyd1BYVjJvYWtVU3ZUckdhZnFzM0M1UkJkUUtZemZ1WGpfU0EwdGxjOEpObmhSQWhVd25pdG00QWg5dVZtRHlLZGtVZ1FCS0ZNVVQyWkRTdTJaT2daVEQ4OW1GcnNZX1JXSEJRMU1BcHlwbDd5ZDAyY3ZzV2dtbG9yeEVZTHNESGs0NGJ1cEkzMk9FZG05eWJjVmFlX2owTFktb3J1Q3IxNEVmdm1XWGFreXFOd204bXNkRkRYZkQ1TURpXy1MQURHNFVFa18wR29jeldlbEZQYUFCbHNhQ1BNRVUycnNEVEwzWnVUM1FCcmNCNzZaUVFEM1BFRlZCQS1FQTNtQnE4RlBrRVFqTF9IY3pUY1FYU2pfbGtsQ1pWd2ZiX21lQzNzaW5LU2w3TG5WVHo0b1F3IiwiZHAiOiJ0TXBUXzlJcVZDb2I2U1k4YUJrU3IyTWwtVXlPQk1TY29wSGRuWExsQ2ZDckVpdXo1YWdLUDVwTzhmNHktbW1nbHE0OEVpanZjY1cyUERHb3lvRzVnN280blIzM3ByOGFmcDlrQWx1QmZmWWpVZXA2WjBFdkVlaEo1RmtmUlA2UVFxbTV3M2ZVcGRPU19wb2RvTDE5bUkzTlVHRTltenpLVm9lamREel92YWMiLCJkcSI6IlpiSXJVdW5HZ2FEQW5rQXhLa0tObUpKdWp3LTJUc0Q5Sk0xUmxIdmZ1XzVxajdRS1dsdEJ4OTBxOWJHOWpXeDJnSHJCeWpPZVdWcXMxZEdUMVhOblVscUpaNWNHTjh4dXpXQmhBLXdjMUhwazhuNXpBNUg3RTlMbzdqa1puOXpaX1IyTEMwNUljS1ppTXBVaHBFUktsR0ZSUGZNb3RRU2cwV002b2laenFRayIsImUiOiJBUUFCIiwia3R5IjoiUlNBIiwibiI6IngxcU4wU0NfT3E1cVh6YjZCWDNuNk1JLWVxem5NZnFwcFVkcTBhWkpMUTB6U0RZSER0OGl4OUp1dGszQTVaOUpiZXJIT3JIdjlvQnNubmZRUnJZRHNLODdyN09hU0dVeHpyRm1ZLTZqRzg4dWZ3VWpBOEdfVl9HaVNYeTQ2VFpZRlpNb2J4UHVjdFJwNmhQeFBUNXQ5a19td2FIYnZ1Vm5zNmYyNnZOVmVrbko2YjhpLWlrbmFxR255VEhNSFNHUmtqX0FuVXlxbGF6cEFZdDZhSGZCQ2lWVllBUERzUUM3ZzYwQ01nNkNPX2hHVDFqRWRZY1c3VDZ6T1FKejE5cURIY0JkODE3dVVZQTEzR2tVaUoycEVJQUM5OTZWOWdKMFoyZGFNa0VVS0wxUlZTZk1mNm9uNURUMkswR0pibEpKVXludER1MUJDMTdBY1RxQ2FvbkdhUSIsInAiOiIzWHFaU0s5WUg0Y2FxajQ1NG9BMlA1X3Z3RTliUFFWRUNkbV9RRnd5T3d5a2tpMEtLbTBpS1NDdWVWbjdvUV92WXdrNWt2S09ZRWJJejMyaU12VGRXMEp0TlJFd0RfVTExQ1dXUGtyOE0tUGpPeDhoUjlTVVlQS0p0d0FGcUNRTWl4dnFMMGJua2I5VnptdDlKUlkxMDBfVk1RSlNJTHZEb0EzbjhzdHhXbGMiLCJxIjoiNW0wZkZfRTJ6SHRkUjBZT3Y5alh4ampycUFzWDRpOHhBNGJXRURFdFI0aUpmVFZlNXM2bU5idU9LenZBLVp3Z2l6ZGE3SGhPQUc4VU9EVEQ2WXcwQmRLVFhkdUQ3LUV2Mkh1ZERRblFOZ1N2a213SFBYZjRIdjdOajRIQVNHdmFzOWYtOGdjTEFEekc1ejlOZ29OVjI4Z2M0ZWt3UVlzQ1NKeTYxYUhPN1Q4IiwicWkiOiJneWVJbWpCMW5KRFE2ZWpmaGJLRGtReFJ1MjlKRC1RemJaSWdZVlkzOHYtN3pyMTl5TDRBSTVCOUMzSzF2VzJibVkwR1VENXRlNUdsQ2lDbHZ6NmpIVzBrRlNIUmY5V3VnSkVSVDQxQ081X25GZ1FMaUxnUXB4VGV3MjEyRVlfajB1aGJIMkhMRnp2blFNcUtNTlcyYmFUMlhFYy10dEFDaE1pb0Q1T0pXalEifQ=="
const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"
const hs256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`

type stubUnixRouter struct {
	routes []adapterPortInUnix.Route
}

func (s stubUnixRouter) Routes() []adapterPortInUnix.Route {
	return s.routes
}

type stubRouteCache struct {
	data map[string]*coreUpstream.UpstreamService
}

type liveRouteCache struct {
	client valkey.Client
	data   map[string]*coreUpstream.UpstreamService
}

type generatedLookupDataset struct {
	domains        int
	pathsPerDomain int
}

func newStubRouteCache(t testing.TB) *stubRouteCache {
	t.Helper()

	var service coreUpstream.UpstreamService
	if err := json.Unmarshal([]byte(unixServerBenchmarkServiceJSON), &service); err != nil {
		t.Fatalf("failed to unmarshal upstream fixture: %v", err)
	}
	service.InitializeResourceIndex()

	return &stubRouteCache{
		data: map[string]*coreUpstream.UpstreamService{
			service.ServiceName: &service,
		},
	}
}

func newGeneratedStubRouteCache(
	t testing.TB,
	serviceName string,
	dataset generatedLookupDataset,
) (*stubRouteCache, string, string) {
	t.Helper()

	service := &coreUpstream.UpstreamService{
		ServiceName: serviceName,
		Resources:   make([]*coreUpstream.UpstreamResource, 0, dataset.domains),
	}

	targetDomain := "api-0.example.com"
	targetRouteSuffix := ((dataset.pathsPerDomain - 1) / 4) * 4

	for domainIndex := 0; domainIndex < dataset.domains; domainIndex++ {
		resource := &coreUpstream.UpstreamResource{
			Domain: fmt.Sprintf("api-%d.example.com", domainIndex),
			Host:   fmt.Sprintf("upstream-%d.internal:8080", domainIndex),
			Paths:  make([]*coreUpstream.UpstreamPath, 0, dataset.pathsPerDomain),
		}

		for pathIndex := 0; pathIndex < dataset.pathsPerDomain; pathIndex++ {
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
			resource.Paths = append(resource.Paths, path)
		}

		service.Resources = append(service.Resources, resource)
	}

	service.InitializeResourceIndex()

	return &stubRouteCache{
			data: map[string]*coreUpstream.UpstreamService{
				service.ServiceName: service,
			},
		},
		targetDomain,
		fmt.Sprintf("/v1/%s/%s/api/0/users/123/posts/%d", serviceName, targetDomain, targetRouteSuffix)
}

func (s *stubRouteCache) LoadCache() {}

func (s *stubRouteCache) Get(key string) (*coreUpstream.UpstreamService, bool) {
	service, ok := s.data[key]
	return service, ok
}

func (s *stubRouteCache) Update(keys []string) error {
	return nil
}

func (s *stubRouteCache) Evict(service string) {
	delete(s.data, service)
}

var _ applicationCache.RouteCache = (*stubRouteCache)(nil)

func (r *liveRouteCache) LoadCache() {
	keys, err := r.keyScan(context.Background(), 0)
	if err != nil {
		panic(err)
	}
	if len(keys) == 0 {
		return
	}

	command := r.client.B().Mget().Key(keys...).Build()
	values, err := r.client.Do(context.Background(), command).AsStrSlice()
	if err != nil {
		panic(err)
	}

	parseRequests := make([]coreUtils.ParseRequest, len(values))
	for idx := range keys {
		serviceName := keys[idx][len("UPSTREAM:"):]
		parseRequests[idx] = coreUtils.NewParseRequest(serviceName, values[idx])
	}

	results := coreUtils.ParseToUpstreamServiceWithInitialize(parseRequests)
	for idx := range results {
		result := results[idx]
		r.data[result.ServiceName] = result
	}
}

func (r *liveRouteCache) Get(key string) (*coreUpstream.UpstreamService, bool) {
	service, ok := r.data[key]
	return service, ok
}

func (r *liveRouteCache) Update(keys []string) error {
	return nil
}

func (r *liveRouteCache) Evict(service string) {
	delete(r.data, service)
}

func (r *liveRouteCache) keyScan(ctx context.Context, cursor uint64) ([]string, error) {
	command := r.client.B().Scan().Cursor(cursor).Match("UPSTREAM:*").Build()
	result, err := r.client.Do(ctx, command).AsScanEntry()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	if result.Cursor > 0 {
		keys, err = r.keyScan(ctx, result.Cursor)
		if err != nil {
			return keys, err
		}
	}

	return append(result.Elements, keys...), nil
}

var _ applicationCache.RouteCache = (*liveRouteCache)(nil)

func newTestUnixServer(socketPath string, routes []adapterPortInUnix.Route) *unixServer {
	cfg := &app_config.AppConfig{}
	cfg.Server.UnixSocketPath = socketPath
	cfg.Server.ReadTimeoutMillisecond = 15
	cfg.Server.WriteTimeoutMillisecond = 25
	cfg.Server.IdleTimeoutMillisecond = 35

	return &unixServer{
		UnixRouterProvider: &adapterPortInUnixDi.UnixRouterProvider{
			UpStreamRouter: stubUnixRouter{routes: routes},
		},
		httpServerProperties: cfg,
	}
}

func newLookupUnixServer(t testing.TB, socketPath string) *unixServer {
	t.Helper()

	useCase := applicationService.NewUpstreamLookupService(newStubRouteCache(t))
	return newLookupUnixServerWithUseCase(socketPath, useCase)
}

func newLookupUnixServerWithRouteCache(socketPath string, routeCache applicationCache.RouteCache) *unixServer {
	useCase := applicationService.NewUpstreamLookupService(routeCache)
	return newLookupUnixServerWithUseCase(socketPath, useCase)
}

func newLookupUnixServerWithUseCase(socketPath string, useCase applicationPortIn.UpstreamLookupUseCase) *unixServer {
	router := adapterPortInUnixHandler.NewUpStreamHandler(useCase)

	cfg := &app_config.AppConfig{}
	cfg.Server.UnixSocketPath = socketPath
	cfg.Server.ReadTimeoutMillisecond = 15
	cfg.Server.WriteTimeoutMillisecond = 25
	cfg.Server.IdleTimeoutMillisecond = 35

	return &unixServer{
		UnixRouterProvider: &adapterPortInUnixDi.UnixRouterProvider{
			UpStreamRouter: router,
		},
		httpServerProperties: cfg,
	}
}

func keyDataForAlgorithm(algorithm string) string {
	switch algorithm {
	case "RS256":
		return rsaJWKBase64
	case "ES256":
		return ecdsaJWKBase64
	case "HS256":
		return base64.StdEncoding.EncodeToString(coreUtils.ToBytesFromString(hs256JWKJSON))
	default:
		panic("unsupported algorithm: " + algorithm)
	}
}

func buildValkeyBackedServiceJSON(serviceName string, algorithm string, withAuth bool) string {
	authorizationBlock := ""
	if withAuth {
		authorizationBlock = fmt.Sprintf(`
		"authorization": {
			"algorithm": %q,
			"key_data": %q,
			"user_key": "user_id"
		},`, algorithm, keyDataForAlgorithm(algorithm))
	}

	return fmt.Sprintf(`{
		"service_name": %q,%s
		"resources": [
			{
				"domain": "api.example.com",
				"host": "upstream-server-1.internal:8080",
				"paths": [
					{
						"path": "/api/users",
						"method": "GET",
						"request_timeout": 5000,
						"response_timeout": 10000,
						"check_authorization": false,
						"cache_timeout": 10
					},
					{
						"path": "/api/users/{userId}/posts/{postId}",
						"method": "GET",
						"request_timeout": 3000,
						"response_timeout": 5000,
						"check_authorization": %t,
						"rate_limit_count": 0
					}
				]
			}
		]
	}`, serviceName, authorizationBlock, withAuth)
}

func loadValkeyMasterHostForTest(t testing.TB) string {
	t.Helper()

	if host := strings.TrimSpace(os.Getenv("VALKEY_MASTER_HOST")); host != "" {
		return host
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current test file path")
	}

	dir := filepath.Dir(currentFile)
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			envMap, err := godotenv.Read(envPath)
			if err != nil {
				t.Fatalf("failed to read .env: %v", err)
			}

			host := strings.TrimSpace(envMap["VALKEY_MASTER_HOST"])
			if host == "" {
				t.Fatalf("VALKEY_MASTER_HOST missing in %s", envPath)
			}
			return host
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Skip("VALKEY_MASTER_HOST not configured; skipping valkey integration tests")
	return ""
}

func lockValkeyTestScope(t testing.TB) {
	t.Helper()

	lockFile, err := os.OpenFile(filepath.Join(os.TempDir(), "gateway-controller-valkey-test.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("failed to open valkey test lock file: %v", err)
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		t.Fatalf("failed to acquire valkey test lock: %v", err)
	}

	t.Cleanup(func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		_ = lockFile.Close()
	})
}

func newLiveRouteCache(t testing.TB, serviceName string) *liveRouteCache {
	return newLiveRouteCacheWithAlgorithm(t, serviceName, "RS256")
}

func newLiveRouteCacheWithAlgorithm(t testing.TB, serviceName string, algorithm string) *liveRouteCache {
	return newLiveRouteCacheWithAuth(t, serviceName, algorithm, true)
}

func newLiveRouteCacheWithoutAuth(t testing.TB, serviceName string) *liveRouteCache {
	return newLiveRouteCacheWithAuth(t, serviceName, "HS256", false)
}

func newLiveRouteCacheWithAuth(t testing.TB, serviceName string, algorithm string, withAuth bool) *liveRouteCache {
	t.Helper()
	lockValkeyTestScope(t)

	appConfig := &app_config.AppConfig{}
	appConfig.Valkey.MasterHost = loadValkeyMasterHostForTest(t)
	client := valkey2.NewValkeyClient(appConfig)
	valkeyClient := client.GetClient()
	t.Cleanup(valkeyClient.Close)
	t.Cleanup(client.GetSubscribeClient().Close)

	serviceJSON := buildValkeyBackedServiceJSON(serviceName, algorithm, withAuth)
	key := "UPSTREAM:" + serviceName
	setCommand := valkeyClient.B().Set().Key(key).Value(serviceJSON).Build()
	if err := valkeyClient.Do(context.Background(), setCommand).Error(); err != nil {
		t.Fatalf("failed to seed valkey route: %v", err)
	}

	t.Cleanup(func() {
		delCommand := valkeyClient.B().Del().Key(key).Build()
		_ = valkeyClient.Do(context.Background(), delCommand).Error()
	})

	routeCache := &liveRouteCache{
		client: valkeyClient,
		data:   make(map[string]*coreUpstream.UpstreamService),
	}
	routeCache.LoadCache()

	return routeCache
}

func newJWTAccessToken(t testing.TB, serviceName string, userID string) string {
	t.Helper()

	codec, err := gjwt.NewCodec(serviceName)
	if err != nil {
		t.Fatalf("failed to build jwt codec: %v", err)
	}

	now := time.Now()
	token, err := codec.Serialize(
		nil,
		func(claims map[string]any) {
			claims["user_id"] = userID
			claims[string(gjwt.IssuedAt)] = now.Unix()
			claims[string(gjwt.Expiration)] = now.Add(time.Hour).Unix()
		},
	)
	if err != nil {
		t.Fatalf("failed to sign jwt token: %v", err)
	}

	return token
}

func deleteValkeyService(t testing.TB, routeCache *liveRouteCache, serviceName string) {
	t.Helper()

	key := "UPSTREAM:" + serviceName
	command := routeCache.client.B().Del().Key(key).Build()
	if err := routeCache.client.Do(context.Background(), command).Error(); err != nil {
		t.Fatalf("failed to delete valkey service: %v", err)
	}
}

func waitForUnixServer(t testing.TB, client *http.Client, url string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("unix server was not ready: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func newUnixSocketHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

func TestUnixServerNewServeMux_RegistersRoutes(t *testing.T) {
	t.Parallel()

	server := newTestUnixServer("", []adapterPortInUnix.Route{
		{
			Method: http.MethodGet,
			Path:   "/v1/upstream",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/upstream", nil)

	server.newServeMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestUnixServerNewHTTPServer_AppliesConfiguredTimeouts(t *testing.T) {
	t.Parallel()

	server := newTestUnixServer("", nil)
	handler := http.NewServeMux()

	httpServer := server.newHTTPServer(handler)

	if httpServer.ReadTimeout != 15*time.Millisecond {
		t.Fatalf("expected read timeout %s, got %s", 15*time.Millisecond, httpServer.ReadTimeout)
	}
	if httpServer.WriteTimeout != 25*time.Millisecond {
		t.Fatalf("expected write timeout %s, got %s", 25*time.Millisecond, httpServer.WriteTimeout)
	}
	if httpServer.IdleTimeout != 35*time.Millisecond {
		t.Fatalf("expected idle timeout %s, got %s", 35*time.Millisecond, httpServer.IdleTimeout)
	}
	if httpServer.Handler == nil {
		t.Fatal("expected handler to be assigned to http server")
	}
}

func TestUnixServerLookupRoute_ReturnsMatchedUpstreamInfo(t *testing.T) {
	t.Parallel()

	server := newLookupUnixServer(t, "")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/upstream?path=/v1/member-api/api.example.com/api/users/123/posts/456",
		nil,
	)

	server.newServeMux().ServeHTTP(recorder, request)

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var result applicationDto.UpStreamInfo
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Host != "upstream-server-1.internal:8080" {
		t.Fatalf("expected host %q, got %q", "upstream-server-1.internal:8080", result.Host)
	}
	if result.Path != "/api/users/123/posts/456" {
		t.Fatalf("expected path %q, got %q", "/api/users/123/posts/456", result.Path)
	}
	if result.OriginalPath != "/api/users/{userId}/posts/{postId}" {
		t.Fatalf("expected original path %q, got %q", "/api/users/{userId}/posts/{postId}", result.OriginalPath)
	}
	if result.Method != http.MethodGet {
		t.Fatalf("expected method %q, got %q", http.MethodGet, result.Method)
	}
	if result.RequestTimeout != 3000 {
		t.Fatalf("expected request timeout %d, got %d", 3000, result.RequestTimeout)
	}
	if result.ResponseTimeout != 5000 {
		t.Fatalf("expected response timeout %d, got %d", 5000, result.ResponseTimeout)
	}
}

func TestUnixServerLookupRoute_ReturnsJSONErrorOnInvalidRequest(t *testing.T) {
	t.Parallel()

	server := newLookupUnixServer(t, "")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/upstream", nil)

	server.newServeMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var result coreAdapterIn.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if result.Message == "" {
		t.Fatal("expected error message to be populated")
	}
}

func TestUnixServerLookupRoute_WithValkeyAndJWT_ReturnsUserKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		algorithm string
	}{
		{name: "RS256", algorithm: "RS256"},
		{name: "ES256", algorithm: "ES256"},
		{name: "HS256", algorithm: "HS256"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceName := "member-api-auth-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAlgorithm(t, serviceName, tc.algorithm)
			token := newJWTAccessToken(t, serviceName, "user-123")
			server := newLookupUnixServerWithRouteCache("", routeCache)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/upstream?path=/v1/"+serviceName+"/api.example.com/api/users/123/posts/456",
				nil,
			)
			request.Header.Set("Authorization", token)

			server.newServeMux().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}

			var result applicationDto.UpStreamInfo
			if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if result.UserKey != "user-123" {
				t.Fatalf("expected user key %q, got %#v", "user-123", result.UserKey)
			}
		})
	}
}

func TestUnixServerLookupRoute_UsesLocalCacheAfterValkeyDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		algorithm string
	}{
		{name: "RS256", algorithm: "RS256"},
		{name: "ES256", algorithm: "ES256"},
		{name: "HS256", algorithm: "HS256"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceName := "member-api-local-route-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAlgorithm(t, serviceName, tc.algorithm)
			token := newJWTAccessToken(t, serviceName, "user-456")
			server := newLookupUnixServerWithRouteCache("", routeCache)

			deleteValkeyService(t, routeCache, serviceName)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/upstream?path=/v1/"+serviceName+"/api.example.com/api/users/123/posts/456",
				nil,
			)
			request.Header.Set("Authorization", token)

			server.newServeMux().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}

			var result applicationDto.UpStreamInfo
			if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if result.Host != "upstream-server-1.internal:8080" {
				t.Fatalf("expected host %q, got %q", "upstream-server-1.internal:8080", result.Host)
			}
			if result.UserKey != "user-456" {
				t.Fatalf("expected user key %q, got %#v", "user-456", result.UserKey)
			}
		})
	}
}

func BenchmarkUnixServerNewServeMux(b *testing.B) {
	routes := make([]adapterPortInUnix.Route, 0, 64)
	for i := range 64 {
		path := "/bench/" + string(rune('a'+(i%26))) + "/" + string(rune('a'+((i/26)%26)))
		routes = append(routes, adapterPortInUnix.Route{
			Method: http.MethodGet,
			Path:   path,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		})
	}

	server := newTestUnixServer("", routes)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = server.newServeMux()
	}
}

func BenchmarkUnixServerLookupRouteOverUnixSocket(b *testing.B) {
	socketPath := filepath.Join(os.TempDir(), "gwc-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
	server := newLookupUnixServer(b, socketPath)

	listener, err := newUnixConnBufferedListener(socketPath, benchmarkUnixSocketBufferBytes, benchmarkUnixSocketBufferBytes)
	if err != nil {
		b.Fatalf("listen unix socket: %v", err)
	}

	httpServer := server.newHTTPServer(server.newServeMux())
	go func() {
		_ = httpServer.Serve(listener)
	}()

	b.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_ = httpServer.Shutdown(shutdownContext)
		_ = os.Remove(socketPath)
	})

	client := newUnixSocketHTTPClient(socketPath)
	b.Cleanup(func() {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	})

	targetURL := "http://unix/v1/upstream?path=/v1/member-api/api.example.com/api/users/123/posts/456"
	waitForUnixServer(b, client, targetURL)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(targetURL)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}

		_, err = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			b.Fatalf("read body failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			b.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	}
}

func BenchmarkUnixServerLookupRouteOverUnixSocketScaled(b *testing.B) {
	cases := []struct {
		name    string
		dataset generatedLookupDataset
	}{
		{
			name: "4-domains-x-32-paths",
			dataset: generatedLookupDataset{
				domains:        4,
				pathsPerDomain: 32,
			},
		},
		{
			name: "16-domains-x-64-paths",
			dataset: generatedLookupDataset{
				domains:        16,
				pathsPerDomain: 64,
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := "member-api-scale-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache, _, targetPath := newGeneratedStubRouteCache(b, serviceName, tc.dataset)

			socketPath := filepath.Join(os.TempDir(), "gwc-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
			server := newLookupUnixServerWithRouteCache(socketPath, routeCache)

			listener, err := newUnixConnBufferedListener(socketPath, benchmarkUnixSocketBufferBytes, benchmarkUnixSocketBufferBytes)
			if err != nil {
				b.Fatalf("listen unix socket: %v", err)
			}

			httpServer := server.newHTTPServer(server.newServeMux())
			go func() {
				_ = httpServer.Serve(listener)
			}()

			b.Cleanup(func() {
				shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_ = httpServer.Shutdown(shutdownContext)
				_ = os.Remove(socketPath)
			})

			client := newUnixSocketHTTPClient(socketPath)
			b.Cleanup(func() {
				if transport, ok := client.Transport.(*http.Transport); ok {
					transport.CloseIdleConnections()
				}
			})

			targetURL := "http://unix/v1/upstream?path=" + targetPath
			waitForUnixServer(b, client, targetURL)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp, err := client.Get(targetURL)
				if err != nil {
					b.Fatalf("request failed: %v", err)
				}

				_, err = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					b.Fatalf("read body failed: %v", err)
				}

				if resp.StatusCode != http.StatusOK {
					b.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}

func BenchmarkUnixServerLookupRouteOverUnixSocketWithValkeyAndJWT(b *testing.B) {
	cases := []struct {
		name       string
		algorithm  string
		withAuth   bool
		targetPath string
	}{
		{name: "NoJWT", algorithm: "HS256", withAuth: false, targetPath: "/api/users"},
		{name: "RS256", algorithm: "RS256", withAuth: true, targetPath: "/api/users/123/posts/456"},
		{name: "ES256", algorithm: "ES256", withAuth: true, targetPath: "/api/users/123/posts/456"},
		{name: "HS256", algorithm: "HS256", withAuth: true, targetPath: "/api/users/123/posts/456"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := "member-api-auth-bench-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAuth(b, serviceName, tc.algorithm, tc.withAuth)

			var token string
			if tc.withAuth {
				token = newJWTAccessToken(b, serviceName, "bench-user")
			}

			socketPath := filepath.Join(os.TempDir(), "gwc-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
			server := newLookupUnixServerWithRouteCache(socketPath, routeCache)

			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				b.Fatalf("listen unix socket: %v", err)
			}

			httpServer := server.newHTTPServer(server.newServeMux())
			go func() {
				_ = httpServer.Serve(listener)
			}()

			b.Cleanup(func() {
				shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_ = httpServer.Shutdown(shutdownContext)
				_ = os.Remove(socketPath)
			})

			client := newUnixSocketHTTPClient(socketPath)
			b.Cleanup(func() {
				if transport, ok := client.Transport.(*http.Transport); ok {
					transport.CloseIdleConnections()
				}
			})

			readinessURL := "http://unix/v1/upstream?path=/v1/" + serviceName + "/api.example.com/api/users"
			waitForUnixServer(b, client, readinessURL)

			targetURL := "http://unix/v1/upstream?path=/v1/" + serviceName + "/api.example.com" + tc.targetPath
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				request, err := http.NewRequest(http.MethodGet, targetURL, nil)
				if err != nil {
					b.Fatalf("create request failed: %v", err)
				}
				if tc.withAuth {
					request.Header.Set("Authorization", token)
				}

				resp, err := client.Do(request)
				if err != nil {
					b.Fatalf("request failed: %v", err)
				}

				_, err = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					b.Fatalf("read body failed: %v", err)
				}

				if resp.StatusCode != http.StatusOK {
					b.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}

func BenchmarkUnixServerLookupRouteOverUnixSocketWithValkeyAndJWTThroughputParallel(b *testing.B) {
	cases := []struct {
		name        string
		algorithm   string
		withAuth    bool
		targetPath  string
		parallelism int
	}{
		{name: "NoJWT/P1", algorithm: "HS256", withAuth: false, targetPath: "/api/users", parallelism: 1},
		{name: "NoJWT/P8", algorithm: "HS256", withAuth: false, targetPath: "/api/users", parallelism: 8},
		{name: "HS256/P1", algorithm: "HS256", withAuth: true, targetPath: "/api/users/123/posts/456", parallelism: 1},
		{name: "HS256/P8", algorithm: "HS256", withAuth: true, targetPath: "/api/users/123/posts/456", parallelism: 8},
		{name: "RS256/P1", algorithm: "RS256", withAuth: true, targetPath: "/api/users/123/posts/456", parallelism: 1},
		{name: "RS256/P8", algorithm: "RS256", withAuth: true, targetPath: "/api/users/123/posts/456", parallelism: 8},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceKey := strings.ReplaceAll(tc.name, "/", "-")
			serviceName := "member-api-http-throughput-" + serviceKey + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAuth(b, serviceName, tc.algorithm, tc.withAuth)

			socketPath := filepath.Join(os.TempDir(), "gwc-http-throughput-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
			server := newLookupUnixServerWithRouteCache(socketPath, routeCache)

			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				b.Fatalf("listen unix socket: %v", err)
			}

			httpServer := server.newHTTPServer(server.newServeMux())
			go func() {
				_ = httpServer.Serve(listener)
			}()

			b.Cleanup(func() {
				shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = httpServer.Shutdown(shutdownContext)
				_ = os.Remove(socketPath)
			})

			client := newUnixSocketHTTPClient(socketPath)
			b.Cleanup(func() {
				if transport, ok := client.Transport.(*http.Transport); ok {
					transport.CloseIdleConnections()
				}
			})

			readinessURL := "http://unix/v1/upstream?path=/v1/" + serviceName + "/api.example.com/api/users"
			waitForUnixServer(b, client, readinessURL)

			targetURL := "http://unix/v1/upstream?path=/v1/" + serviceName + "/api.example.com" + tc.targetPath
			var token string
			if tc.withAuth {
				token = newJWTAccessToken(b, serviceName, "bench-user")
			}

			var failed atomic.Bool
			var failureMessage atomic.Value

			b.SetParallelism(tc.parallelism)
			b.ReportAllocs()
			b.ResetTimer()
			start := time.Now()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if failed.Load() {
						continue
					}

					request, err := http.NewRequest(http.MethodGet, targetURL, nil)
					if err != nil {
						failureMessage.Store(fmt.Sprintf("create request failed: %v", err))
						failed.Store(true)
						continue
					}
					if tc.withAuth {
						request.Header.Set("Authorization", token)
					}

					resp, err := client.Do(request)
					if err != nil {
						failureMessage.Store(fmt.Sprintf("request failed: %v", err))
						failed.Store(true)
						continue
					}

					_, err = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					if err != nil {
						failureMessage.Store(fmt.Sprintf("read body failed: %v", err))
						failed.Store(true)
						continue
					}
					if resp.StatusCode != http.StatusOK {
						failureMessage.Store(fmt.Sprintf("expected status %d, got %d", http.StatusOK, resp.StatusCode))
						failed.Store(true)
						continue
					}
				}
			})
			elapsed := time.Since(start)
			b.ReportMetric(float64(b.N)/elapsed.Seconds(), "req/s")

			if failed.Load() {
				message, _ := failureMessage.Load().(string)
				b.Fatal(message)
			}
		})
	}
}
