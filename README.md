# go-api-archway

`go-api-archway`는 라우팅 정책 관리(control plane)와 요청 프록시(data plane)를 분리한 API Gateway 프로젝트입니다.

---

## 1. 핵심 구성

- **gateway**: 외부 HTTP 요청 수신, 정책 lookup 호출, upstream reverse proxy 수행
- **gateway-controller**: Valkey route를 메모리에 로드하고 Unix socket lookup API 제공
- **core**: 라우팅/파싱/JWT/에러 등 공통 Go 모듈
- **console**: PostgreSQL 기반 정책 관리 + Valkey projection(backoffice)

---

## 2. 아키텍처

```mermaid
flowchart LR
    Client[Client] -->|HTTP :80| Gateway[gateway]
    Gateway -->|HTTP over Unix socket| Controller[gateway-controller]
    Controller -->|startup SCAN/MGET| Valkey[(Valkey)]
    Valkey -->|ROUTE_OPERATIONS pub/sub| Controller
    Controller -->|upstream metadata| Gateway
    Gateway -->|Reverse Proxy| Upstream[Upstream Service]

    Operator[Operator] --> ConsoleFront[console front]
    ConsoleFront --> ConsoleBackend[console backend]
    ConsoleBackend --> PostgreSQL[(PostgreSQL)]
    ConsoleBackend -->|UPSTREAM:* projection| Valkey
```

핫패스에서 gateway가 Valkey를 직접 조회하지 않고, gateway-controller의 메모리 캐시 lookup 결과를 사용합니다.

---

## 3. 요청 URI 규칙 (Gateway Contract)

요청 경로 형식:

```text
/{service}/{version}/{domain}/{resource...}
```

예시:

```text
/member/v1/sessions/token
/member/v1/user/info
```

- `service`: 서비스 식별자
- `version`: API 버전
- `domain`: DDD 도메인 식별자(호스트 아님)
- `resource...`: 도메인 하위 리소스 경로

`domain`은 upstream hostname이 아니라 route 선택 키입니다. 실제 upstream 주소는 route JSON의 `host` 필드로 결정합니다.

---

## 4. Route 모델

Valkey key: `UPSTREAM:{service-name}`

```json
{
  "service_name": "member",
  "authorization": {
    "algorithm": "HS256",
    "key_data": "<base64 encoded JWK>",
    "user_key": "user_id"
  },
  "resources": [
    {
      "domain": "user",
      "host": "127.0.0.1:18081",
      "paths": [
        {
          "path": "/info",
          "method": "GET",
          "request_timeout": 3000,
          "response_timeout": 5000,
          "check_authorization": true,
          "cache_timeout": 0
        }
      ]
    }
  ]
}
```

매칭 규칙:

- path는 정적/동적(`{id}`) 세그먼트를 Trie 기반으로 매칭
- 도메인 일치 실패 시 `domain: ""` fallback resource 조회
- `check_authorization=true`면 JWT verify 수행
- verify 성공 시 `authorization.user_key` claim을 `X-USER` 헤더로 upstream 전달

지원 JWT 알고리즘: `HS256`, `RS256`, `ES256`

---

## 5. Route 동기화

- Source of truth: PostgreSQL(console backend)
- Valkey projection key: `UPSTREAM:{service-name}`
- Pub/Sub channel: `ROUTE_OPERATIONS` (`ADD`, `UPDATE`, `DELETE`)

gateway-controller 동작:

1. 시작 시 `UPSTREAM:*` 대상 `SCAN + MGET` 로드
2. 런타임에는 `ROUTE_OPERATIONS` 이벤트를 구독해 메모리 캐시 갱신

---

## 6. 저장소 구조

```text
.
├── archway/
│   ├── gateway/                 # HTTP reverse proxy
│   ├── gateway-controller/      # Unix socket lookup server
│   ├── .docker/                 # 통합 이미지
│   └── Makefile
├── core/                        # 공통 Go 모듈
├── console/
│   ├── front/                   # React + Vite
│   └── backend/                 # Express + PostgreSQL + Valkey
└── README.md
```

---

## 7. 요구사항

- Go `1.26.1`
- Wire `v0.7.0`
- Docker
- Valkey
- Node.js / npm (console)
- PostgreSQL (console backend)

---

## 8. 빌드

### 8.1 Gateway + Controller

```bash
cd archway
make build
```

산출물:

```text
archway/.build/gateway
archway/.build/gateway-controller
```

Wire만 재생성:

```bash
cd archway
make wire
```

### 8.2 Console

```bash
cd console
npm install
npm run build
```

---

## 9. 런타임 설정

### 9.1 gateway

| Environment variable | Default | Description |
|---|---:|---|
| `GATEWAY_CONTROLLER_UPSTREAM_LOOKUP_BASE_URL` | `http://unix/v1/upstream?path=` | controller lookup URL |
| `GATEWAY_CONTROLLER_SERVER_NETWORK` | `unix` | controller network |
| `GATEWAY_CONTROLLER_UNIX_SOCKET_PATH` | `/tmp/gateway-controller.sock` | controller Unix socket |
| `HTTP_CLIENT_MAX_IDLE_CONNS` | `250` | HTTP idle connections |
| `HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST` | `500` | host별 idle connections |
| `HTTP_CLIENT_IDLE_CONN_TIMEOUT` | `90` | idle timeout(sec) |
| `HTTP_CLIENT_TIMEOUT_MILLISECONDS` | `5000` | controller request timeout(ms) |

gateway listen 주소는 현재 `:80`입니다.

### 9.2 gateway-controller

| Environment variable | Default | Description |
|---|---:|---|
| `UNIX_SOCKET_PATH` | `/tmp/gateway-controller.sock` | Unix socket path |
| `READ_TIMEOUT_MILLISECOND` | `10` | read timeout(ms) |
| `WRITE_TIMEOUT_MILLISECOND` | `10` | write timeout(ms) |
| `IDLE_TIMEOUT_MILLISECOND` | `120` | idle timeout(ms) |
| `VALKEY_MASTER_HOST` | required | Valkey master |
| `VALKEY_REPLICA_HOSTS` | empty | replicas (comma-separated) |
| `VALKEY_READFROM` | `master` | read policy |

`.env`는 선택 사항이며 실제 환경변수만으로 실행 가능합니다.

### 9.3 console backend

| Environment variable | Default | Description |
|---|---:|---|
| `PORT` | `8080` | backend listen port |
| `POSTGRES_CONNECTION_STRING` | required | PostgreSQL connection string |
| `VALKEY_URL` | required | Valkey URL |
| `OUTBOX_POLL_INTERVAL_MS` | `3000` | outbox polling interval |
| `OUTBOX_BATCH_SIZE` | `10` | outbox batch size |

---

## 10. 로컬 실행

Valkey `127.0.0.1:6379` 기준.

### 10.1 gateway-controller

```bash
cd archway/gateway-controller
VALKEY_MASTER_HOST=127.0.0.1:6379 go run ./cmd
```

### 10.2 gateway

```bash
cd archway/gateway
go run ./cmd
```

gateway-controller가 먼저 올라와 Unix socket을 생성해야 정상 lookup이 됩니다.

---

## 11. E2E 빠른 검증

### 11.1 임시 upstream 실행

```bash
python3 -u -c '
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
class H(BaseHTTPRequestHandler):
    def do_GET(self):
        body=b"ok"
        self.send_response(200)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *args): pass
ThreadingHTTPServer(("127.0.0.1", 18081), H).serve_forever()
'
```

### 11.2 route 등록

```bash
docker exec valkey-master valkey-cli SET 'UPSTREAM:member' \
'{"service_name":"member","resources":[{"domain":"user","host":"127.0.0.1:18081","paths":[{"path":"/info","method":"GET","request_timeout":3000,"response_timeout":5000,"check_authorization":false,"cache_timeout":0}]}]}'
```

### 11.3 요청 검증

```bash
curl -i 'http://127.0.0.1/member/v1/user/info'
```

---

## 12. 테스트 명령

### 12.1 Go (gateway)

```bash
cd archway/gateway
go test ./...
```

### 12.2 Go (gateway-controller)

```bash
cd archway/gateway-controller
go test ./...
```

### 12.3 Console

```bash
cd console/backend
npm run test

cd ../
npm run build
```

---

## 13. 성능 벤치마크 요약

측정 환경:

- Apple M4, Darwin arm64
- Go `1.26.1`
- Valkey `8.0.2`
- localhost loopback, HTTP keep-alive
- 경로: `gateway -> unix socket -> gateway-controller -> gateway -> upstream`

### 13.1 JWT Codec

| Algorithm | Operation | Time | Memory | Allocations |
|---|---|---:|---:|---:|
| HS256 | sign | 1.20µs | 2,114 B/op | 35 allocs/op |
| HS256 | verify | 1.69µs | 2,392 B/op | 48 allocs/op |
| RS256 | sign | 856µs | 8,621 B/op | 76 allocs/op |
| RS256 | verify | 23.3µs | 3,640 B/op | 53 allocs/op |
| ES256 | sign | 16.7µs | 8,531 B/op | 97 allocs/op |
| ES256 | verify | 38.0µs | 3,288 B/op | 65 allocs/op |

### 13.2 Gateway End-to-End (Concurrency 32, No JWT)

| Run | Throughput | Average | P50 | P95 | P99 | Max |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 39.8K req/s | 793µs | 699µs | 1.47ms | 1.99ms | 6.31ms |
| 2 | 40.8K req/s | 774µs | 678µs | 1.43ms | 1.95ms | 4.89ms |
| 3 | 38.9K req/s | 817µs | 704µs | 1.60ms | 2.16ms | 4.80ms |

### 13.3 CPU / RSS (Concurrency 32, 15s sustained)

| Scenario | Sustained throughput | Gateway CPU | Controller CPU | Combined CPU | Combined RSS |
|---|---:|---:|---:|---:|---:|
| No JWT | 42.2K req/s | 4.46 cores | 1.13 cores | 5.59 cores | 53 MiB |
| HS256 | 36.8K req/s | 4.11 cores | 1.28 cores | 5.39 cores | 57.1 MiB |
| RS256 | 31.8K req/s | 3.53 cores | 2.49 cores | 6.02 cores | 57.5 MiB |
| ES256 | 31.6K req/s | 3.07 cores | 3.31 cores | 6.38 cores | 59.1 MiB |

---

## 14. 100K req/s Pod 산정

가정:

- gateway + gateway-controller를 동일 Pod에 배치
- Pod CPU request `3.5`, limit `4`
- CPU headroom `25%`
- 산식: `RequiredCPU = CombinedCPU × (100K / Throughput) × 1.25`
- 여기서 `RequiredCPU`는 **Pod 1대가 아니라 전체 클러스터 총 CPU**입니다.
- Pod 1대 기준 체감은 `CPU per Pod at min pods = RequiredCPU / RequiredPods(min)`으로 봅니다.

| Scenario | Total CPU for 100K (cluster) | Required Pods (min) | CPU per Pod at min pods | Recommended Pods (+1 spare) |
|---|---:|---:|---:|---:|
| No JWT | 16.6 cores | 5 | 3.32 cores/pod | 6 |
| HS256 | 18.3 cores | 6 | 3.05 cores/pod | 7 |
| RS256 | 23.7 cores | 7 | 3.39 cores/pod | 8 |
| ES256 | 25.3 cores | 8 | 3.16 cores/pod | 9 |

메모리는 실측 RSS 기준으로 `53–59 MiB`이므로, 현재 프로파일에서는 Pod당 `request 256Mi / limit 512Mi`로 시작 가능하며, route 수/키 수/sidecar 증가 시 재측정이 필요합니다.

---

## 15. Docker / Compose 실행

사전 준비: `docker compose` 설치

```bash
docker compose version
```

서비스 포트:

| Service | Port |
|---|---:|
| archway-gateway | `80` |
| console-frontend | `8081` |
| console-backend | `8080` |
| PostgreSQL | `5431` |
| Valkey master | `6379` |
| Valkey replica | `6380` |

### 15.1 Infra compose 실행

```bash
cd .doker
docker compose up -d
```

### 15.2 Gateway 앱 build/run

```bash
docker build \
  -f archway/.docker/Dockerfile \
  -t go-api-archway:latest \
  .
```

```bash
docker run --rm --name archway-gateway \
  -p 80:80 \
  -e VALKEY_MASTER_HOST=host.docker.internal:6379 \
  go-api-archway:latest
```

### 15.3 Console 앱 build/run

```bash
cd console
POSTGRES_CONNECTION_STRING=postgres://postgres:postgres@127.0.0.1:5431/postgres \
VALKEY_URL=redis://127.0.0.1:6379 \
npm run docker:build
```

```bash
cd console
POSTGRES_CONNECTION_STRING=postgres://postgres:postgres@127.0.0.1:5431/postgres \
VALKEY_URL=redis://127.0.0.1:6379 \
npm run docker:run
```

---

## 16. 운영 체크리스트

- gateway-controller Unix socket 준비 후 gateway 시작
- route 등록 후 controller가 로드/갱신했는지 확인
- HPA는 CPU(60~70%) + latency(P95) + controller CPU를 함께 기준화
- service mesh/TLS 사용 시 CPU/latency 여유 추가
- 민감정보(JWT key/token/DB credential) 로그/저장소 노출 금지
