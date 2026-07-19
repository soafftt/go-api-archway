package cache

import (
	"context"
	"core/gjwt"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gateway/controller/adapter/config/valkey"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

const rsaJWKBase64 = "eyJkIjoiS3V3SjFjbWRzc1pEMjkxTDVxdk5wMVZQTzMySHg4Q0hEekY1dzhyd1BYVjJvYWtVU3ZUckdhZnFzM0M1UkJkUUtZemZ1WGpfU0EwdGxjOEpObmhSQWhVd25pdG00QWg5dVZtRHlLZGtVZ1FCS0ZNVVQyWkRTdTJaT2daVEQ4OW1GcnNZX1JXSEJRMU1BcHlwbDd5ZDAyY3ZzV2dtbG9yeEVZTHNESGs0NGJ1cEkzMk9FZG05eWJjVmFlX2owTFktb3J1Q3IxNEVmdm1XWGFreXFOd204bXNkRkRYZkQ1TURpXy1MQURHNFVFa18wR29jeldlbEZQYUFCbHNhQ1BNRVUycnNEVEwzWnVUM1FCcmNCNzZaUVFEM1BFRlZCQS1FQTNtQnE4RlBrRVFqTF9IY3pUY1FYU2pfbGtsQ1pWd2ZiX21lQzNzaW5LU2w3TG5WVHo0b1F3IiwiZHAiOiJ0TXBUXzlJcVZDb2I2U1k4YUJrU3IyTWwtVXlPQk1TY29wSGRuWExsQ2ZDckVpdXo1YWdLUDVwTzhmNHktbW1nbHE0OEVpanZjY1cyUERHb3lvRzVnN280blIzM3ByOGFmcDlrQWx1QmZmWWpVZXA2WjBFdkVlaEo1RmtmUlA2UVFxbTV3M2ZVcGRPU19wb2RvTDE5bUkzTlVHRTltenpLVm9lamREel92YWMiLCJkcSI6IlpiSXJVdW5HZ2FEQW5rQXhLa0tObUpKdWp3LTJUc0Q5Sk0xUmxIdmZ1XzVxajdRS1dsdEJ4OTBxOWJHOWpXeDJnSHJCeWpPZVdWcXMxZEdUMVhOblVscUpaNWNHTjh4dXpXQmhBLXdjMUhwazhuNXpBNUg3RTlMbzdqa1puOXpaX1IyTEMwNUljS1ppTXBVaHBFUktsR0ZSUGZNb3RRU2cwV002b2laenFRayIsImUiOiJBUUFCIiwia3R5IjoiUlNBIiwibiI6IngxcU4wU0NfT3E1cVh6YjZCWDNuNk1JLWVxem5NZnFwcFVkcTBhWkpMUTB6U0RZSER0OGl4OUp1dGszQTVaOUpiZXJIT3JIdjlvQnNubmZRUnJZRHNLODdyN09hU0dVeHpyRm1ZLTZqRzg4dWZ3VWpBOEdfVl9HaVNYeTQ2VFpZRlpNb2J4UHVjdFJwNmhQeFBUNXQ5a19td2FIYnZ1Vm5zNmYyNnZOVmVrbko2YjhpLWlrbmFxR255VEhNSFNHUmtqX0FuVXlxbGF6cEFZdDZhSGZCQ2lWVllBUERzUUM3ZzYwQ01nNkNPX2hHVDFqRWRZY1c3VDZ6T1FKejE5cURIY0JkODE3dVVZQTEzR2tVaUoycEVJQUM5OTZWOWdKMFoyZGFNa0VVS0wxUlZTZk1mNm9uNURUMkswR0pibEpKVXludER1MUJDMTdBY1RxQ2FvbkdhUSIsInAiOiIzWHFaU0s5WUg0Y2FxajQ1NG9BMlA1X3Z3RTliUFFWRUNkbV9RRnd5T3d5a2tpMEtLbTBpS1NDdWVWbjdvUV92WXdrNWt2S09ZRWJJejMyaU12VGRXMEp0TlJFd0RfVTExQ1dXUGtyOE0tUGpPeDhoUjlTVVlQS0p0d0FGcUNRTWl4dnFMMGJua2I5VnptdDlKUlkxMDBfVk1RSlNJTHZEb0EzbjhzdHhXbGMiLCJxIjoiNW0wZkZfRTJ6SHRkUjBZT3Y5alh4ampycUFzWDRpOHhBNGJXRURFdFI0aUpmVFZlNXM2bU5idU9LenZBLVp3Z2l6ZGE3SGhPQUc4VU9EVEQ2WXcwQmRLVFhkdUQ3LUV2Mkh1ZERRblFOZ1N2a213SFBYZjRIdjdOajRIQVNHdmFzOWYtOGdjTEFEekc1ejlOZ29OVjI4Z2M0ZWt3UVlzQ1NKeTYxYUhPN1Q4IiwicWkiOiJneWVJbWpCMW5KRFE2ZWpmaGJLRGtReFJ1MjlKRC1RemJaSWdZVlkzOHYtN3pyMTl5TDRBSTVCOUMzSzF2VzJibVkwR1VENXRlNUdsQ2lDbHZ6NmpIVzBrRlNIUmY5V3VnSkVSVDQxQ081X25GZ1FMaUxnUXB4VGV3MjEyRVlfajB1aGJIMkhMRnp2blFNcUtNTlcyYmFUMlhFYy10dEFDaE1pb0Q1T0pXalEifQ=="
const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"
const hs256JWKJSON = `{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`

type generatedRouteDataset struct {
	serviceCount   int
	domains        int
	pathsPerDomain int
	withAuth       bool
}

func keyDataForAlgorithm(algorithm string) string {
	switch algorithm {
	case "RS256":
		return rsaJWKBase64
	case "ES256":
		return ecdsaJWKBase64
	case "HS256":
		return base64.StdEncoding.EncodeToString([]byte(hs256JWKJSON))
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

func buildGeneratedValkeyServiceJSON(serviceName string, algorithm string, dataset generatedRouteDataset) string {
	payload := map[string]any{
		"service_name": serviceName,
	}

	if dataset.withAuth {
		payload["authorization"] = map[string]any{
			"algorithm": algorithm,
			"key_data":  keyDataForAlgorithm(algorithm),
			"user_key":  "user_id",
		}
	}

	resources := make([]map[string]any, 0, dataset.domains)
	for domainIndex := 0; domainIndex < dataset.domains; domainIndex++ {
		paths := make([]map[string]any, 0, dataset.pathsPerDomain)
		for pathIndex := 0; pathIndex < dataset.pathsPerDomain; pathIndex++ {
			path := fmt.Sprintf("/api/%d/items/%d", domainIndex, pathIndex)
			if pathIndex%3 == 0 {
				path = fmt.Sprintf("/api/%d/users/{userId}/orders/%d", domainIndex, pathIndex)
			}

			paths = append(paths, map[string]any{
				"path":                path,
				"method":              "GET",
				"request_timeout":     3000 + pathIndex,
				"response_timeout":    5000 + pathIndex,
				"check_authorization": dataset.withAuth && domainIndex == 0 && pathIndex == dataset.pathsPerDomain-1,
				"cache_timeout":       pathIndex % 10,
				"rate_limit_count":    0,
			})
		}

		resources = append(resources, map[string]any{
			"domain": fmt.Sprintf("api-%d.example.com", domainIndex),
			"host":   fmt.Sprintf("upstream-%d.internal:8080", domainIndex),
			"paths":  paths,
		})
	}

	payload["resources"] = resources

	raw, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("marshal generated valkey service json: %w", err))
	}

	return string(raw)
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

	t.Fatal("failed to locate .env with VALKEY_MASTER_HOST")
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

func seedValkeyService(t testing.TB, serviceName string, algorithm string) valkey.ValkeyClient {
	t.Helper()
	lockValkeyTestScope(t)

	appConfig := &app_config.AppConfig{}
	appConfig.Valkey.MasterHost = loadValkeyMasterHostForTest(t)
	client := valkey.NewValkeyClient(appConfig)
	valkeyClient := client.GetClient()
	t.Cleanup(valkeyClient.Close)
	t.Cleanup(client.GetSubscribeClient().Close)

	key := "UPSTREAM:" + serviceName
	value := buildValkeyBackedServiceJSON(serviceName, algorithm, true)
	command := valkeyClient.B().Set().Key(key).Value(value).Build()
	if err := valkeyClient.Do(context.Background(), command).Error(); err != nil {
		t.Fatalf("failed to seed valkey service: %v", err)
	}

	t.Cleanup(func() {
		command := valkeyClient.B().Del().Key(key).Build()
		_ = valkeyClient.Do(context.Background(), command).Error()
	})

	return client
}

func deleteValkeyService(t testing.TB, client valkey.ValkeyClient, serviceName string) {
	t.Helper()

	valkeyClient := client.GetClient()
	key := "UPSTREAM:" + serviceName
	command := valkeyClient.B().Del().Key(key).Build()
	if err := valkeyClient.Do(context.Background(), command).Error(); err != nil {
		t.Fatalf("failed to delete valkey service: %v", err)
	}
}

func seedGeneratedValkeyServices(
	t testing.TB,
	dataset generatedRouteDataset,
	servicePrefix string,
	algorithm string,
) (valkey.ValkeyClient, []string) {
	t.Helper()
	lockValkeyTestScope(t)

	appConfig := &app_config.AppConfig{}
	appConfig.Valkey.MasterHost = loadValkeyMasterHostForTest(t)
	client := valkey.NewValkeyClient(appConfig)
	valkeyClient := client.GetClient()
	t.Cleanup(valkeyClient.Close)
	t.Cleanup(client.GetSubscribeClient().Close)

	serviceNames := make([]string, 0, dataset.serviceCount)
	keys := make([]string, 0, dataset.serviceCount)
	for index := 0; index < dataset.serviceCount; index++ {
		serviceName := fmt.Sprintf("%s-%d", servicePrefix, index)
		key := "UPSTREAM:" + serviceName
		value := buildGeneratedValkeyServiceJSON(serviceName, algorithm, dataset)
		command := valkeyClient.B().Set().Key(key).Value(value).Build()
		if err := valkeyClient.Do(context.Background(), command).Error(); err != nil {
			t.Fatalf("failed to seed generated valkey service %q: %v", serviceName, err)
		}
		serviceNames = append(serviceNames, serviceName)
		keys = append(keys, key)
	}

	t.Cleanup(func() {
		if len(keys) == 0 {
			return
		}
		command := valkeyClient.B().Del().Key(keys...).Build()
		_ = valkeyClient.Do(context.Background(), command).Error()
	})

	return client, serviceNames
}

func TestRouteValkeyCacheLoadCache_LoadsServiceAndRegistersJWTKey(t *testing.T) {
	t.Parallel()

	for _, algorithm := range []string{"RS256", "ES256", "HS256"} {
		t.Run(algorithm, func(t *testing.T) {
			serviceName := "member-api-valkey-" + algorithm + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			client := seedValkeyService(t, serviceName, algorithm)
			routeCache := NewRouteValkeyCache(client)

			routeCache.LoadCache()

			service, ok := routeCache.Get(serviceName)
			if !ok {
				t.Fatalf("expected service %q to be loaded from valkey", serviceName)
			}
			if service == nil {
				t.Fatal("expected loaded service to be non-nil")
			}
			if service.Authorization == nil {
				t.Fatal("expected authorization metadata to be loaded")
			}
			if !gjwt.HasKey(serviceName) {
				t.Fatalf("expected jwt key %q to be registered from valkey payload", serviceName)
			}
		})
	}
}

func TestRouteValkeyCacheGet_UsesLocalCacheAfterValkeyDelete(t *testing.T) {
	t.Parallel()

	for _, algorithm := range []string{"RS256", "ES256", "HS256"} {
		t.Run(algorithm, func(t *testing.T) {
			serviceName := "member-api-local-cache-" + algorithm + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			client := seedValkeyService(t, serviceName, algorithm)
			routeCache := NewRouteValkeyCache(client)

			routeCache.LoadCache()
			deleteValkeyService(t, client, serviceName)

			service, ok := routeCache.Get(serviceName)
			if !ok {
				t.Fatalf("expected service %q to remain in local cache", serviceName)
			}
			if service == nil {
				t.Fatal("expected cached service to be non-nil")
			}
			if service.ServiceName != serviceName {
				t.Fatalf("expected cached service name %q, got %q", serviceName, service.ServiceName)
			}
		})
	}
}

func TestRouteValkeyCacheLoadCache_LoadsAllScanPages(t *testing.T) {
	t.Parallel()

	dataset := generatedRouteDataset{
		serviceCount:   20,
		domains:        1,
		pathsPerDomain: 1,
		withAuth:       false,
	}
	servicePrefix := "multi-page-scan-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	client, serviceNames := seedGeneratedValkeyServices(t, dataset, servicePrefix, "HS256")
	routeCache := NewRouteValkeyCache(client)

	routeCache.LoadCache()

	for _, serviceName := range serviceNames {
		if _, ok := routeCache.Get(serviceName); !ok {
			t.Fatalf("expected service %q to be loaded from valkey", serviceName)
		}
	}
}

func BenchmarkRouteValkeyCacheLoadCache(b *testing.B) {
	cases := []struct {
		name      string
		algorithm string
		withAuth  bool
	}{
		{name: "NoJWT", algorithm: "HS256", withAuth: false},
		{name: "RS256", algorithm: "RS256", withAuth: true},
		{name: "ES256", algorithm: "ES256", withAuth: true},
		{name: "HS256", algorithm: "HS256", withAuth: true},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := "member-api-load-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			client := seedValkeyService(b, serviceName, tc.algorithm)
			if !tc.withAuth {
				valkeyClient := client.GetClient()
				key := "UPSTREAM:" + serviceName
				value := buildValkeyBackedServiceJSON(serviceName, tc.algorithm, false)
				command := valkeyClient.B().Set().Key(key).Value(value).Build()
				if err := valkeyClient.Do(context.Background(), command).Error(); err != nil {
					b.Fatalf("failed to override valkey service without auth: %v", err)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				routeCache := NewRouteValkeyCache(client)
				routeCache.LoadCache()

				if _, ok := routeCache.Get(serviceName); !ok {
					b.Fatalf("expected service %q to be loaded from valkey", serviceName)
				}
			}
		})
	}
}

func BenchmarkRouteValkeyCacheLoadCacheScaled(b *testing.B) {
	cases := []struct {
		name      string
		algorithm string
		dataset   generatedRouteDataset
	}{
		{
			name:      "10-services-x-4-domains-x-16-paths",
			algorithm: "HS256",
			dataset: generatedRouteDataset{
				serviceCount:   10,
				domains:        4,
				pathsPerDomain: 16,
				withAuth:       true,
			},
		},
		{
			name:      "50-services-x-8-domains-x-32-paths",
			algorithm: "HS256",
			dataset: generatedRouteDataset{
				serviceCount:   50,
				domains:        8,
				pathsPerDomain: 32,
				withAuth:       true,
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			servicePrefix := "generated-load-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			client, serviceNames := seedGeneratedValkeyServices(b, tc.dataset, servicePrefix, tc.algorithm)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				routeCache := NewRouteValkeyCache(client)
				routeCache.LoadCache()

				if _, ok := routeCache.Get(serviceNames[0]); !ok {
					b.Fatalf("expected service %q to be loaded from valkey", serviceNames[0])
				}
				if _, ok := routeCache.Get(serviceNames[len(serviceNames)-1]); !ok {
					b.Fatalf("expected service %q to be loaded from valkey", serviceNames[len(serviceNames)-1])
				}
			}
		})
	}
}
