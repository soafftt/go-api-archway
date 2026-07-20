package grpc_server

import (
	"context"
	coreUpstream "core/domain/upstream"
	"core/gjwt"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/controller/adapter/config/app_config"
	"gateway/controller/adapter/config/server"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	adapterPortInGrpcDi "gateway/controller/adapter/port/in/grpc/di"
	adapterPortInGrpcHandler "gateway/controller/adapter/port/in/grpc/handler"
	applicationPortIn "gateway/controller/application/port/in"
	applicationDto "gateway/controller/application/port/in/dto"
	applicationCache "gateway/controller/application/port/out/cache"
	applicationService "gateway/controller/application/service"
	pb "gateway/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const grpcBufConnSize = 1024 * 1024

type stubUpstreamLookupUseCase struct {
	result applicationDto.UpStreamLookupResult
}

type stubListenerServer struct{}

func (stubListenerServer) Start() {}

func applyDefaultGrpcServerConfig(cfg *app_config.AppConfig) {
	cfg.Server.Grpc.Network = "unix"
	cfg.Server.Grpc.MaxRecvMsgBytes = 4 * 1024 * 1024
	cfg.Server.Grpc.MaxSendMsgBytes = 10 * 1024 * 1024
	cfg.Server.Grpc.ReadBufferBytes = 32 * 1024
	cfg.Server.Grpc.WriteBufferBytes = 32 * 1024
	cfg.Server.Grpc.ConnectionTimeoutMillisecond = 5000
	cfg.Server.Grpc.MaxConcurrentStreams = 2048
	cfg.Server.Grpc.NumStreamWorkers = 0
	cfg.Server.Grpc.KeepaliveMaxConnectionIdleMs = 15 * 60 * 1000
	cfg.Server.Grpc.KeepaliveMaxConnectionAgeMs = 30 * 60 * 1000
	cfg.Server.Grpc.KeepaliveTimeMs = 2 * 60 * 1000
	cfg.Server.Grpc.KeepaliveTimeoutMs = 20 * 1000
	cfg.Server.Grpc.KeepaliveEnforcementMinTimeMs = 20 * 1000
	cfg.Server.Grpc.PermitWithoutStream = true
	cfg.Server.Grpc.GracefulStopTimeoutMillisecond = 300
}

func (s stubUpstreamLookupUseCase) LookUpFromRequest(_ applicationDto.UpStreamLookupRequest) applicationDto.UpStreamLookupResult {
	return s.result
}

func (s stubUpstreamLookupUseCase) Lookup(
	_ string,
	_ string,
	_ string,
	_ string,
	_ *string,
) applicationDto.UpStreamLookupResult {
	return s.result
}

func buildGrpcServiceJSON(serviceName string, algorithm string, withAuth bool) string {
	if withAuth {
		return fmt.Sprintf(`{
	"service_name": %q,
	"authorization": {
		"algorithm": %q,
		"key_data": %q,
		"user_key": "user_id"
	},
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
					"check_authorization": true,
					"rate_limit_count": 0
				}
			]
		}
	]
}`, serviceName, algorithm, keyDataForAlgorithm(algorithm))
	}

	return buildValkeyBackedServiceJSON(serviceName, algorithm, false)
}

func newStubRouteCacheWithAuth(t testing.TB, serviceName string, algorithm string, withAuth bool) *stubRouteCache {
	t.Helper()

	serviceJSON := buildGrpcServiceJSON(serviceName, algorithm, withAuth)
	service := &coreUpstream.UpstreamService{}
	if err := json.Unmarshal([]byte(serviceJSON), service); err != nil {
		t.Fatalf("failed to parse service json: %v", err)
	}
	service.InitializeResourceIndex()
	if withAuth {
		if err := gjwt.RegisterKeyByString(serviceName, keyDataForAlgorithm(algorithm), gjwt.JSONKey, algorithm); err != nil {
			t.Fatalf("failed to register jwt key: %v", err)
		}
	}

	return &stubRouteCache{
		data: map[string]*coreUpstream.UpstreamService{
			serviceName: service,
		},
	}
}

func newTestGrpcServer(routeCache applicationCache.RouteCache) *grpcServer {
	useCase := applicationService.NewUpstreamLookupService(routeCache)
	controller := adapterPortInGrpcHandler.NewUpstreamLookupController(useCase)
	registrars := adapterPortInGrpcDi.NewGrpcServiceRegistrars(controller)

	cfg := &app_config.AppConfig{}
	applyDefaultGrpcServerConfig(cfg)

	return &grpcServer{
		GrpcServiceProvider: &adapterPortInGrpcDi.GrpcServiceProvider{Registrars: registrars},
		AppConfig:           cfg,
	}
}

func newBufconnGrpcClient(
	t testing.TB,
	server *grpcServer,
) (pb.UpstreamLookupServiceClient, *grpc.ClientConn, func()) {
	t.Helper()

	listener := bufconn.Listen(grpcBufConnSize)
	grpcServer := server.newServer()
	server.registerServices(grpcServer)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}

	connection, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn grpc server: %v", err)
	}

	cleanup := func() {
		_ = connection.Close()
		grpcServer.Stop()
		_ = listener.Close()
	}

	return pb.NewUpstreamLookupServiceClient(connection), connection, cleanup
}

func newUnixSocketGrpcClient(
	t testing.TB,
	server *grpcServer,
	socketPath string,
) (pb.UpstreamLookupServiceClient, *grpc.ClientConn, func()) {
	t.Helper()

	_ = os.Remove(socketPath)

	listener, err := newUnixConnBufferedListener(socketPath, benchmarkUnixSocketBufferBytes, benchmarkUnixSocketBufferBytes)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}

	grpcServer := server.newServer()
	server.registerServices(grpcServer)
	go func() {
		_ = grpcServer.Serve(listener)
	}()

	connection, err := grpc.NewClient(
		"passthrough:///unix-grpc",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcServer.Stop()
		_ = listener.Close()
		t.Fatalf("failed to dial unix grpc server: %v", err)
	}

	cleanup := func() {
		_ = connection.Close()
		grpcServer.Stop()
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}

	return pb.NewUpstreamLookupServiceClient(connection), connection, cleanup
}

func newTCPGrpcClient(
	t testing.TB,
	server *grpcServer,
) (pb.UpstreamLookupServiceClient, *grpc.ClientConn, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}

	grpcServer := server.newServer()
	server.registerServices(grpcServer)
	go func() {
		_ = grpcServer.Serve(listener)
	}()

	connection, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcServer.Stop()
		_ = listener.Close()
		t.Fatalf("failed to dial tcp grpc server: %v", err)
	}

	cleanup := func() {
		_ = connection.Close()
		grpcServer.Stop()
		_ = listener.Close()
	}

	return pb.NewUpstreamLookupServiceClient(connection), connection, cleanup
}

func TestGrpcServerLookup_ReturnsMatchedUpstreamInfo(t *testing.T) {
	t.Parallel()

	serviceName := "member-api-grpc-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	server := newTestGrpcServer(newStubRouteCacheWithAuth(t, serviceName, "HS256", false))
	client, _, cleanup := newBufconnGrpcClient(t, server)
	defer cleanup()

	response, err := client.Lookup(context.Background(), &pb.UpstreamLookupRequest{
		Path: "/v1/" + serviceName + "/api.example.com/api/users",
	})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if response.GetError() != "" {
		t.Fatalf("expected empty business error, got %q", response.GetError())
	}
	if response.GetInfo() == nil {
		t.Fatal("expected upstream info")
	}
	if response.GetInfo().GetHost() != "upstream-server-1.internal:8080" {
		t.Fatalf("expected host %q, got %q", "upstream-server-1.internal:8080", response.GetInfo().GetHost())
	}
}

func TestNewListenerServer_SelectsByTransport(t *testing.T) {
	t.Parallel()

	appConfig := &app_config.AppConfig{}
	unixServer := stubListenerServer{}
	grpcServer := &grpcServer{}

	appConfig.Server.Transport = ""
	selected := server.NewListenerServer(appConfig, unixServer, grpcServer)
	if selected != unixServer {
		t.Fatal("expected unix server for default transport")
	}

	appConfig.Server.Transport = "grpc"
	selected = server.NewListenerServer(appConfig, unixServer, grpcServer)
	if selected != grpcServer {
		t.Fatal("expected grpc server for grpc transport")
	}
}

func TestGrpcServerLookup_BusinessErrorInResponse(t *testing.T) {
	t.Parallel()

	useCase := stubUpstreamLookupUseCase{result: applicationDto.NewErrUpStreamLookupResult(errors.New("service not found"))}
	controller := adapterPortInGrpcHandler.NewUpstreamLookupController(useCase)
	registrars := adapterPortInGrpcDi.NewGrpcServiceRegistrars(controller)
	server := &grpcServer{
		GrpcServiceProvider: &adapterPortInGrpcDi.GrpcServiceProvider{Registrars: registrars},
		AppConfig:           &app_config.AppConfig{},
	}
	applyDefaultGrpcServerConfig(server.AppConfig)

	client, _, cleanup := newBufconnGrpcClient(t, server)
	defer cleanup()

	response, err := client.Lookup(context.Background(), &pb.UpstreamLookupRequest{})
	if err != nil {
		t.Fatalf("expected nil transport error, got %v", err)
	}
	if response.GetError() == "" {
		t.Fatal("expected business error in response field")
	}
}

func TestGrpcServerLookup_WithJWT_ReturnsStringUserKey(t *testing.T) {
	t.Parallel()

	serviceName := "member-api-grpc-jwt-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	server := newTestGrpcServer(newStubRouteCacheWithAuth(t, serviceName, "HS256", true))
	token := newJWTAccessToken(t, serviceName, "grpc-user")
	client, _, cleanup := newBufconnGrpcClient(t, server)
	defer cleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
	response, err := client.Lookup(ctx, &pb.UpstreamLookupRequest{
		Path: "/v1/" + serviceName + "/api.example.com/api/users/123/posts/456",
	})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if response.GetError() != "" {
		t.Fatalf("expected empty business error, got %q", response.GetError())
	}

	if response.GetInfo().GetUserKey() != "grpc-user" {
		t.Fatalf("expected user key %q, got %q", "grpc-user", response.GetInfo().GetUserKey())
	}
}

func benchmarkGrpcLookup(b *testing.B, algorithm string, withAuth bool) {
	serviceName := "member-api-grpc-bench-" + algorithm + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	server := newTestGrpcServer(newStubRouteCacheWithAuth(b, serviceName, algorithm, withAuth))
	client, _, cleanup := newBufconnGrpcClient(b, server)
	defer cleanup()

	request := &pb.UpstreamLookupRequest{
		Path: "/v1/" + serviceName + "/api.example.com/api/users",
	}

	requestContext := context.Background()
	if withAuth {
		request.Path = "/v1/" + serviceName + "/api.example.com/api/users/123/posts/456"
		token := newJWTAccessToken(b, serviceName, "bench-user")
		requestContext = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response, err := client.Lookup(requestContext, request)
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
		if response.GetError() != "" {
			b.Fatalf("unexpected business error: %s", response.GetError())
		}
		if withAuth && response.GetInfo().GetUserKey() == "" {
			b.Fatal("expected user key in auth benchmark")
		}
	}
}

func BenchmarkGrpcLookup_NoJWT(b *testing.B) {
	benchmarkGrpcLookup(b, "HS256", false)
}

func BenchmarkGrpcLookup_HS256(b *testing.B) {
	benchmarkGrpcLookup(b, "HS256", true)
}

func BenchmarkGrpcLookup_RS256(b *testing.B) {
	benchmarkGrpcLookup(b, "RS256", true)
}

func BenchmarkGrpcLookupOverUnixSocketWithValkeyAndJWT(b *testing.B) {
	cases := []struct {
		name      string
		algorithm string
		withAuth  bool
		path      string
	}{
		{
			name:      "NoJWT",
			algorithm: "HS256",
			withAuth:  false,
			path:      "api.example.com/api/users",
		},
		{
			name:      "HS256",
			algorithm: "HS256",
			withAuth:  true,
			path:      "api.example.com/api/users/123/posts/456",
		},
		{
			name:      "RS256",
			algorithm: "RS256",
			withAuth:  true,
			path:      "api.example.com/api/users/123/posts/456",
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := "member-api-grpc-auth-bench-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAuth(b, serviceName, tc.algorithm, tc.withAuth)
			server := newTestGrpcServer(routeCache)

			socketPath := filepath.Join(os.TempDir(), "gwc-grpc-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
			client, _, cleanup := newUnixSocketGrpcClient(b, server, socketPath)
			defer cleanup()

			var token string
			if tc.withAuth {
				token = newJWTAccessToken(b, serviceName, "bench-user")
			}

			request := &pb.UpstreamLookupRequest{
				Path: "/v1/" + serviceName + "/" + tc.path,
			}
			requestContext := context.Background()
			if tc.withAuth {
				requestContext = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
			}

			readinessDeadline := time.Now().Add(2 * time.Second)
			for {
				response, err := client.Lookup(requestContext, request)
				if err == nil && response.GetError() == "" {
					break
				}
				if time.Now().After(readinessDeadline) {
					b.Fatalf("grpc unix server was not ready: err=%v response_error=%v", err, response.GetError())
				}
				time.Sleep(10 * time.Millisecond)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response, err := client.Lookup(requestContext, request)
				if err != nil {
					b.Fatalf("lookup failed: %v", err)
				}
				if response.GetError() != "" {
					b.Fatalf("unexpected business error: %s", response.GetError())
				}
				if tc.withAuth && response.GetInfo().GetUserKey() == "" {
					b.Fatal("expected user key in auth benchmark")
				}
			}
		})
	}
}

func BenchmarkGrpcLookupOverTCPWithValkeyAndJWT(b *testing.B) {
	cases := []struct {
		name      string
		algorithm string
		withAuth  bool
		path      string
	}{
		{
			name:      "NoJWT",
			algorithm: "HS256",
			withAuth:  false,
			path:      "api.example.com/api/users",
		},
		{
			name:      "HS256",
			algorithm: "HS256",
			withAuth:  true,
			path:      "api.example.com/api/users/123/posts/456",
		},
		{
			name:      "RS256",
			algorithm: "RS256",
			withAuth:  true,
			path:      "api.example.com/api/users/123/posts/456",
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceName := "member-api-grpc-tcp-bench-" + tc.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAuth(b, serviceName, tc.algorithm, tc.withAuth)
			server := newTestGrpcServer(routeCache)
			client, _, cleanup := newTCPGrpcClient(b, server)
			defer cleanup()

			var token string
			if tc.withAuth {
				token = newJWTAccessToken(b, serviceName, "bench-user")
			}

			request := &pb.UpstreamLookupRequest{
				Path: "/v1/" + serviceName + "/" + tc.path,
			}
			requestContext := context.Background()
			if tc.withAuth {
				requestContext = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
			}

			readinessDeadline := time.Now().Add(2 * time.Second)
			for {
				response, err := client.Lookup(requestContext, request)
				if err == nil && response.GetError() == "" {
					break
				}
				if time.Now().After(readinessDeadline) {
					b.Fatalf("grpc tcp server was not ready: err=%v response_error=%v", err, response.GetError())
				}
				time.Sleep(10 * time.Millisecond)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response, err := client.Lookup(requestContext, request)
				if err != nil {
					b.Fatalf("lookup failed: %v", err)
				}
				if response.GetError() != "" {
					b.Fatalf("unexpected business error: %s", response.GetError())
				}
				if tc.withAuth && response.GetInfo().GetUserKey() == "" {
					b.Fatal("expected user key in auth benchmark")
				}
			}
		})
	}
}

func BenchmarkGrpcLookupOverUnixSocketWithValkeyAndJWTThroughputParallel(b *testing.B) {
	cases := []struct {
		name        string
		algorithm   string
		withAuth    bool
		path        string
		parallelism int
	}{
		{name: "NoJWT/P1", algorithm: "HS256", withAuth: false, path: "api.example.com/api/users", parallelism: 1},
		{name: "NoJWT/P8", algorithm: "HS256", withAuth: false, path: "api.example.com/api/users", parallelism: 8},
		{name: "HS256/P1", algorithm: "HS256", withAuth: true, path: "api.example.com/api/users/123/posts/456", parallelism: 1},
		{name: "HS256/P8", algorithm: "HS256", withAuth: true, path: "api.example.com/api/users/123/posts/456", parallelism: 8},
		{name: "RS256/P1", algorithm: "RS256", withAuth: true, path: "api.example.com/api/users/123/posts/456", parallelism: 1},
		{name: "RS256/P8", algorithm: "RS256", withAuth: true, path: "api.example.com/api/users/123/posts/456", parallelism: 8},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			serviceKey := strings.ReplaceAll(tc.name, "/", "-")
			serviceName := "member-api-grpc-throughput-" + serviceKey + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			routeCache := newLiveRouteCacheWithAuth(b, serviceName, tc.algorithm, tc.withAuth)
			server := newTestGrpcServer(routeCache)

			socketPath := filepath.Join(os.TempDir(), "gwc-grpc-throughput-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")
			client, _, cleanup := newUnixSocketGrpcClient(b, server, socketPath)
			defer cleanup()

			request := &pb.UpstreamLookupRequest{
				Path: "/v1/" + serviceName + "/" + tc.path,
			}
			requestContext := context.Background()
			if tc.withAuth {
				token := newJWTAccessToken(b, serviceName, "bench-user")
				requestContext = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
			}

			readinessDeadline := time.Now().Add(2 * time.Second)
			for {
				response, err := client.Lookup(requestContext, request)
				if err == nil && response.GetError() == "" {
					break
				}
				if time.Now().After(readinessDeadline) {
					b.Fatalf("grpc unix server was not ready: err=%v response_error=%v", err, response.GetError())
				}
				time.Sleep(10 * time.Millisecond)
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
					response, err := client.Lookup(requestContext, request)
					if err != nil {
						failureMessage.Store(fmt.Sprintf("lookup failed: %v", err))
						failed.Store(true)
						continue
					}
					if response.GetError() != "" {
						failureMessage.Store(fmt.Sprintf("unexpected business error: %s", response.GetError()))
						failed.Store(true)
						continue
					}
					if tc.withAuth && response.GetInfo().GetUserKey() == "" {
						failureMessage.Store("expected user key in auth benchmark")
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

var _ applicationPortIn.UpstreamLookupUseCase = (*stubUpstreamLookupUseCase)(nil)
