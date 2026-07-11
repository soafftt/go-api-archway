# console backoffice

`console`은 `gateway-controller`의 라우팅 규칙을 관리하는 backoffice workspace이다.  
프론트엔드는 운영자가 규칙을 편집하는 UI를 제공하고, 백엔드는 규칙을 검증한 뒤 PostgreSQL에 저장하고 Valkey projection 및 pub/sub를 갱신한다.

## 목표

- gateway path 규칙을 backoffice에서 생성/수정/삭제
- `gateway-controller`가 소비하는 `UPSTREAM:{service}` snapshot JSON 유지
- 규칙 변경 시 `ROUTE_OPERATIONS` 채널로 publish
- 프론트/백엔드 프로세스를 분리하여 동일 서버에서 함께 운영

## 디렉터리 구조

```text
console/
  front/               # React + Vite + Tailwind UI
  backend/             # TypeScript + Express + PostgreSQL + Valkey API
  package.json         # npm workspace
  README.md
```

## 아키텍처

### 1. Frontend (`console/front`)

- React + TypeScript + Vite
- Tailwind CSS 기반 UI
- 기본 포트: `8081`
- `/api` 요청을 `127.0.0.1:8080`으로 proxy

주요 화면/기능:

- service 목록
- service/resource/path 편집
- authorization 설정
- client validation
- gateway path sample 확인
- 저장된 규칙 기준 rewrite preview
- Valkey snapshot preview

### 2. Backend (`console/backend`)

- Express + TypeScript
- 기본 포트: `8080`
- PostgreSQL을 source of truth로 사용
- Valkey를 `gateway-controller` 호환 projection 저장소로 사용
- outbox worker가 PostgreSQL 변경분을 Valkey에 반영하고 publish

주요 API:

- `GET /api/v1/upstream-services`
- `GET /api/v1/upstream-services/:serviceName`
- `POST /api/v1/upstream-services`
- `PUT /api/v1/upstream-services/:serviceName`
- `DELETE /api/v1/upstream-services/:serviceName`
- `POST /api/v1/upstream-services/preview-match`
- `POST /api/v1/upstream-services/:serviceName/republish`
- `GET /health`

### 3. Storage / Sync

#### PostgreSQL

다음 테이블을 사용한다.

- `upstream_services`
- `upstream_resources`
- `upstream_paths`
- `route_change_outbox`

스키마는 `backend/sql/schema.sql`에 있다.

#### Valkey

- key prefix: `UPSTREAM:`
- channel: `ROUTE_OPERATIONS`

projection payload는 Go `gateway-controller`가 읽는 snake_case JSON 형식과 맞춘다.

#### gateway-controller 연동

이번 작업에서 `archway/gateway-controller`도 같이 보완했다.

- 시작 시 Valkey route cache `LoadCache()`
- `ROUTE_OPERATIONS` subscribe
- `ADD/UPDATE/DELETE` 이벤트를 받아 메모리 cache 동기화
- `authorization.userKey` claim 이름을 실제 JWT 파싱에 반영

즉, backoffice 변경이 실제 gateway lookup 결과까지 전파된다.

## 런타임 설정

### Backend env

```bash
POSTGRES_CONNECTION_STRING=postgresql://postgres:1234@127.0.0.1:5431/postgres
VALKEY_URL=redis://127.0.0.1:6379
PORT=8080
OUTBOX_POLL_INTERVAL_MS=3000
OUTBOX_BATCH_SIZE=10
```

### Frontend

- 별도 env 없이 Vite dev server 사용
- 기본 포트 `8081`

### gateway-controller

```bash
VALKEY_HOSTS=127.0.0.1:6379,127.0.0.1:6380
```

## 실행 절차

### 1. 의존성 설치

```bash
cd console
npm install
```

### 2. PostgreSQL 스키마 초기화

```bash
cd console
POSTGRES_CONNECTION_STRING=postgresql://postgres:1234@127.0.0.1:5431/postgres \
npm run db:init --workspace @console/backend
```

### 3. Backend 실행

```bash
cd console
POSTGRES_CONNECTION_STRING=postgresql://postgres:1234@127.0.0.1:5431/postgres \
VALKEY_URL=redis://127.0.0.1:6379 \
PORT=8080 \
npm run dev --workspace @console/backend
```

### 4. Frontend 실행

```bash
cd console
npm run dev --workspace @console/front
```

### 5. gateway-controller 실행

```bash
cd archway/gateway-controller
VALKEY_HOSTS=127.0.0.1:6379 VALKEY_REPLICA_HOSTS=127.0.0.1:6380 go run ./cmd
```

## Docker 실행

`console/.docker`에 프론트/백엔드 Dockerfile을 분리했고, 빌드 산출물은 `console/.build/front`, `console/.build/backend`로 각각 패키징된다.

```bash
cd console
POSTGRES_CONNECTION_STRING=******127.0.0.1:5431/postgres \
VALKEY_URL=redis://127.0.0.1:6379 \
npm run docker:build
```

```bash
cd console
POSTGRES_CONNECTION_STRING=******127.0.0.1:5431/postgres \
VALKEY_URL=redis://127.0.0.1:6379 \
npm run docker:run
```

## 이번 작업 절차

이번 구현은 아래 순서로 진행됐다.

1. `console/.github/agent.md` 확인
2. 기존 Go 도메인/Valkey 저장 규칙 조사
3. frontend / backend 설계 에이전트로 분리 설계 수행
4. 초기 스캐폴드 구현
5. PostgreSQL 전환
6. frontend / backend 역할 기반 리뷰 수행
7. 리뷰 결과 반영
8. gateway-controller 연동 보완
9. 실환경 포트 기준 live validation 수행

## 수행한 검토/리뷰

### 설계 검토

- frontend 설계 리뷰 후 UX gap 수정
- backend 설계 리뷰 후 PostgreSQL/outbox/sync gap 수정
- 최종적으로 frontend / backend 모두 초기 마일스톤 기준 승인

### 코드리뷰

frontend 리뷰에서 반영한 항목:

- 생성 후 잘못된 service 선택 문제 수정
- stale fetch 응답이 editor를 덮어쓰는 문제 수정
- preview 동작 범위를 저장된 규칙 기준으로 명확화
- 잘못된 sample path 생성 보정

backend 리뷰에서 반영한 항목:

- PostgreSQL unique violation → 409 처리
- outbox stale processing 복구
- bootstrap schema init script 추가
- gateway-controller live sync 연결
- cache update 구현 및 fatal 종료 제거
- subscription readiness 대기 처리 보강
- JWT claim 누락/비문자열 검증 추가

## 검증 절차

### console 검증

```bash
cd console
npm run build

cd console/backend
npm run test
```

### gateway-controller 검증

```bash
cd archway/gateway-controller
go test ./adapter/port/out/cache ./adapter/config ./cmd
```

### live validation

다음을 실제로 확인했다.

1. backend `/health` 응답
2. frontend `http://localhost:8081` 응답
3. gateway-controller unix socket 생성
4. backend `POST`
5. gateway-controller lookup 결과 생성 반영
6. backend `PUT`
7. gateway-controller lookup 결과 갱신 반영
8. backend `DELETE`
9. gateway-controller lookup 결과 삭제 반영
10. Valkey `UPSTREAM:{service}` snapshot 반영
11. `ROUTE_OPERATIONS` publish 확인

## 현재 운영 기준

- frontend: `http://localhost:8081`
- backend: `http://console-backend:8080` (compose 네트워크 내부)
- PostgreSQL: `127.0.0.1:5431`
- Valkey: `127.0.0.1:6379` (master), `127.0.0.1:6380` (replica)
- gateway-controller socket: `/tmp/gateway-controller.sock`

## 현재 알려진 비차단 사항

- `preview-match`는 실제 controller와 동일하게 path 중심 lookup을 보여주고, 저장된 method는 metadata로만 표시한다.
- outbox / route listener에 대한 운영용 metrics, health detail, 문서화는 더 보강할 여지가 있다.
- frontend는 초기 마일스톤 수준의 UX이며, 더 정교한 form field-level validation/UI test는 후속 확장 대상이다.
