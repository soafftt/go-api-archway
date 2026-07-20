package in_test

import (
	"bufio"
	"fmt"
	"gateway/adapter/config"
	"gateway/adapter/config/client"
	ingateway "gateway/adapter/in"
	"gateway/adapter/in/middleware"
	"gateway/adapter/out/ratelimit"
	"gateway/application/service"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGatewayE2E_RewritePathVariableViaGatewayController(t *testing.T) {
	t.Parallel()

	valkeyAddress := "127.0.0.1:6379"
	serviceName := "gateway-e2e-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	valkeyKey := "UPSTREAM:" + serviceName
	socketPath := filepath.Join(os.TempDir(), "gateway-controller-e2e-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".sock")

	upstreamPathCh := make(chan string, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamPathCh <- request.URL.Path
		_, _ = writer.Write([]byte("ok"))
	}))
	defer upstream.Close()

	payload := buildGatewayE2EValkeyPayload(serviceName, upstream.URL)
	if err := valkeySet(valkeyAddress, valkeyKey, payload); err != nil {
		t.Fatalf("failed to set valkey fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = valkeyDel(valkeyAddress, valkeyKey)
	})

	cmd := exec.Command("go", "run", "./cmd")
	cmd.Dir = filepath.Clean(filepath.Join("..", "..", "..", "gateway-controller"))
	cmd.Env = append(os.Environ(),
		"VALKEY_MASTER_HOST="+valkeyAddress,
		"UNIX_SOCKET_PATH="+socketPath,
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start gateway-controller: %v", err)
	}
	t.Cleanup(func() {
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

	waitForGatewayControllerSocket(t, socketPath)

	appConfig := &config.AppConfig{}
	appConfig.ClientNetworkConfig.BaseURL = "http://unix/v1/upstream?path="
	appConfig.ClientNetworkConfig.Network = "unix"
	appConfig.ClientNetworkConfig.UnixSocketPath = socketPath
	appConfig.HttpClient.TimeoutMilliSeconds = 3000

	gatewayControllerClient := client.NewHttpClient(appConfig)
	upstreamLookupPort := controlplane.NewUpstreamLookup(appConfig, gatewayControllerClient)
	upstreamLookupUseCase := service.NewUpstreamLookupService(upstreamLookupPort)
	requestMiddleware := middleware.NewRequestMiddleware(upstreamLookupUseCase, ratelimit.NewRateLimit())

	handler := middleware.Chain(ingateway.NewGatewayProxy().HttpProxy, requestMiddleware)

	request := httptest.NewRequest(
		http.MethodGet,
		"http://gateway.local/v1/"+serviceName+"/api.example.com/echo/user-123/posts/777?trace=true",
		nil,
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%q", http.StatusOK, response.Code, response.Body.String())
	}

	select {
	case gotPath := <-upstreamPathCh:
		if gotPath != "/echo/user-123/posts/777" {
			t.Fatalf("expected rewritten upstream path %q, got %q", "/echo/user-123/posts/777", gotPath)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("upstream did not receive request")
	}
}

func waitForGatewayControllerSocket(t *testing.T, socketPath string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("gateway-controller unix socket not ready: %s", socketPath)
}

func buildGatewayE2EValkeyPayload(serviceName, upstreamHost string) string {
	return fmt.Sprintf(`{
	"service_name": %q,
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
					"check_authorization": false,
					"cache_timeout": 0,
					"rate_limit_count": 0
				}
			]
		}
	]
}`, serviceName, upstreamHost)
}

func valkeySet(addr, key, value string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)
	if _, err := conn.Write([]byte(command)); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "+OK") {
		return fmt.Errorf("unexpected valkey SET response: %s", strings.TrimSpace(line))
	}

	return nil
}

func valkeyDel(addr, key string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	command := fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
	if _, err := conn.Write([]byte(command)); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	_, err = reader.ReadString('\n')
	return err
}
