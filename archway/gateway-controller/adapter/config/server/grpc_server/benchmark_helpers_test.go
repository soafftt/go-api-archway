package grpc_server

import (
	"context"
	coreUpstream "core/domain/upstream"
	"core/gjwt"
	coreUtils "core/utils"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"gateway/controller/adapter/config/app_config"
	valkeyConfig "gateway/controller/adapter/config/valkey"

	"github.com/joho/godotenv"
	"github.com/valkey-io/valkey-go"
)

const benchmarkUnixSocketBufferBytes = 1 << 20 // 1 MiB

const rsaJWKBase64 = "eyJkIjoiS3V3SjFjbWRzc1pEMjkxTDVxdk5wMVZQTzMySHg4Q0hEekY1dzhyd1BYVjJvYWtVU3ZUckdhZnFzM0M1UkJkUUtZemZ1WGpfU0EwdGxjOEpObmhSQWhVd25pdG00QWg5dVZtRHlLZGtVZ1FCS0ZNVVQyWkRTdTJaT2daVEQ4OW1GcnNZX1JXSEJRMU1BcHlwbDd5ZDAyY3ZzV2dtbG9yeEVZTHNESGs0NGJ1cEkzMk9FZG05eWJjVmFlX2owTFktb3J1Q3IxNEVmdm1XWGFreXFOd204bXNkRkRYZkQ1TURpXy1MQURHNFVFa18wR29jeldlbEZQYUFCbHNhQ1BNRVUycnNEVEwzWnVUM1FCcmNCNzZaUVFEM1BFRlZCQS1FQTNtQnE4RlBrRVFqTF9IY3pUY1FYU2pfbGtsQ1pWd2ZiX21lQzNzaW5LU2w3TG5WVHo0b1F3IiwiZHAiOiJ0TXBUXzlJcVZDb2I2U1k4YUJrU3IyTWwtVXlPQk1TY29wSGRuWExsQ2ZDckVpdXo1YWdLUDVwTzhmNHktbW1nbHE0OEVpanZjY1cyUERHb3lvRzVnN280blIzM3ByOGFmcDlrQWx1QmZmWWpVZXA2WjBFdkVlaEo1RmtmUlA2UVFxbTV3M2ZVcGRPU19wb2RvTDE5bUkzTlVHRTltenpLVm9lamREel92YWMiLCJkcSI6IlpiSXJVdW5HZ2FEQW5rQXhLa0tObUpKdWp3LTJUc0Q5Sk0xUmxIdmZ1XzVxajdRS1dsdEJ4OTBxOWJHOWpXeDJnSHJCeWpPZVdWcXMxZEdUMVhOblVscUpaNWNHTjh4dXpXQmhBLXdjMUhwazhuNXpBNUg3RTlMbzdqa1puOXpaX1IyTEMwNUljS1ppTXBVaHBFUktsR0ZSUGZNb3RRU2cwV002b2laenFRayIsImUiOiJBUUFCIiwia3R5IjoiUlNBIiwibiI6IngxcU4wU0NfT3E1cVh6YjZCWDNuNk1JLWVxem5NZnFwcFVkcTBhWkpMUTB6U0RZSER0OGl4OUp1dGszQTVaOUpiZXJIT3JIdjlvQnNubmZRUnJZRHNLODdyN09hU0dVeHpyRm1ZLTZqRzg4dWZ3VWpBOEdfVl9HaVNYeTQ2VFpZRlpNb2J4UHVjdFJwNmhQeFBUNXQ5a19td2FIYnZ1Vm5zNmYyNnZOVmVrbko2YjhpLWlrbmFxR255VEhNSFNHUmtqX0FuVXlxbGF6cEFZdDZhSGZCQ2lWVllBUERzUUM3ZzYwQ01nNkNPX2hHVDFqRWRZY1c3VDZ6T1FKejE5cURIY0JkODE3dVVZQTEzR2tVaUoycEVJQUM5OTZWOWdKMFoyZGFNa0VVS0wxUlZTZk1mNm9uNURUMkswR0pibEpKVXludER1MUJDMTdBY1RxQ2FvbkdhUSIsInAiOiIzWHFaU0s5WUg0Y2FxajQ1NG9BMlA1X3Z3RTliUFFWRUNkbV9RRnd5T3d5a2tpMEtLbTBpS1NDdWVWbjdvUV92WXdrNWt2S09ZRWJJejMyaU12VGRXMEp0TlJFd0RfVTExQ1dXUGtyOE0tUGpPeDhoUjlTVVlQS0p0d0FGcUNRTWl4dnFMMGJua2I5VnptdDlKUlkxMDBfVk1RSlNJTHZEb0EzbjhzdHhXbGMiLCJxIjoiNW0wZkZfRTJ6SHRkUjBZT3Y5alh4ampycUFzWDRpOHhBNGJXRURFdFI0aUpmVFZlNXM2bU5idU9LenZBLVp3Z2l6ZGE3SGhPQUc4VU9EVEQ2WXcwQmRLVFhkdUQ3LUV2Mkh1ZERRblFOZ1N2a213SFBYZjRIdjdOajRIQVNHdmFzOWYtOGdjTEFEekc1ejlOZ29OVjI4Z2M0ZWt3UVlzQ1NKeTYxYUhPN1Q4IiwicWkiOiJneWVJbWpCMW5KRFE2ZWpmaGJLRGtReFJ1MjlKRC1RemJaSWdZVlkzOHYtN3pyMTl5TDRBSTVCOUMzSzF2VzJibVkwR1VENXRlNUdsQ2lDbHZ6NmpIVzBrRlNIUmY5V3VnSkVSVDQxQ081X25GZ1FMaUxnUXB4VGV3MjEyRVlfajB1aGJIMkhMRnp2blFNcUtNTlcyYmFUMlhFYy10dEFDaE1pb0Q1T0pXalEifQ=="
const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"
const hs256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`

type stubRouteCache struct {
	data map[string]*coreUpstream.UpstreamService
}

func (s *stubRouteCache) LoadCache() {}
func (s *stubRouteCache) Get(key string) (*coreUpstream.UpstreamService, bool) {
	service, ok := s.data[key]
	return service, ok
}
func (s *stubRouteCache) Update(keys []string) error { return nil }
func (s *stubRouteCache) Evict(service string)       { delete(s.data, service) }

type liveRouteCache struct {
	client valkey.Client
	data   map[string]*coreUpstream.UpstreamService
}

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
func (r *liveRouteCache) Update(keys []string) error { return nil }
func (r *liveRouteCache) Evict(service string)       { delete(r.data, service) }

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

type unixConnBufferedListener struct {
	net.Listener
	readBufferBytes  int
	writeBufferBytes int
}

func newUnixConnBufferedListener(socketPath string, readBufferBytes, writeBufferBytes int) (net.Listener, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &unixConnBufferedListener{
		Listener:         listener,
		readBufferBytes:  readBufferBytes,
		writeBufferBytes: writeBufferBytes,
	}, nil
}

func (l *unixConnBufferedListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if unixConn, ok := conn.(*net.UnixConn); ok {
		_ = unixConn.SetReadBuffer(l.readBufferBytes)
		_ = unixConn.SetWriteBuffer(l.writeBufferBytes)
	}
	return conn, nil
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

func newLiveRouteCacheWithAuth(t testing.TB, serviceName string, algorithm string, withAuth bool) *liveRouteCache {
	t.Helper()
	lockValkeyTestScope(t)

	cfg := &app_config.AppConfig{}
	cfg.Valkey.MasterHost = loadValkeyMasterHostForTest(t)
	client := valkeyConfig.NewValkeyClient(cfg)
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
	token, err := codec.Serialize(nil, func(claims map[string]any) {
		claims["user_id"] = userID
		claims[string(gjwt.IssuedAt)] = now.Unix()
		claims[string(gjwt.Expiration)] = now.Add(time.Hour).Unix()
	})
	if err != nil {
		t.Fatalf("failed to sign jwt token: %v", err)
	}
	return token
}
