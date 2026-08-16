# 최적화 관점 수정 필요 영역 정리

## 목적
- 현재 코드 정리 과정에서 **실제 처리량/할당량에 영향을 주는 수정 후보**를 빠르게 식별하기 위한 메모다.
- 기준은 `.docs/benchmark/readout-report.md` 와 핫패스 코드(`gateway`, `gateway-controller`, `core`)다.

## 우선순위 요약
| 우선순위 | 영역 | 주요 파일 | 근거 |
|---|---|---|---|
| P1 | JWT 검증 핫패스 재사용성 부족 | `archway/gateway-controller/application/service/upstream_lookup_service.go` | RS256 구간에서 controller lookup 비용이 크게 상승하며, 인증 경로에서 요청마다 `gjwt.NewCodec`를 생성한다. |
| P1 | 경로 파싱 로직 중복 + 문자열 분할/결합 반복 | `archway/gateway-controller/adapter/port/in/grpc/handler/upstream_lookup_controller.go`, `archway/gateway-controller/application/port/in/dto/upstream_lookup_request.go`, `archway/gateway-controller/utils/path_parser.go` | 동일 규칙 파싱이 여러 곳에 분산되어 있고 `strings.Split/Join`이 반복된다. lookup 요청 수가 많을수록 누적 비용이 커진다. |
| P1 | Trie 검색 시 요청당 map/문자열 재가공 발생 | `core/domain/upstream/path_router.go` | full-chain E2E가 `186~205 allocs/op` 수준이며, 라우터가 매칭마다 `map[string]string`, `copyPathParams`, `strings.ReplaceAll`을 수행한다. |
| P2 | 캐시 초기화 파서의 전역 worker 구조 | `core/utils/upstream_parser.go`, `archway/gateway-controller/adapter/port/out/cache/valkey_route_cache.go` | 시작 시 `SCAN + MGET + JSON parse` 흐름에서 전역 50개 goroutine, 전역 채널, 요청당 `[]byte` 변환이 있다. |
| P2 | gateway ↔ controller 호출에서 context 전파 부재 | `archway/gateway/adapter/out/controlplane/grpc/upstream_lookup.go` | 요청 취소/타임아웃이 controller RPC로 전파되지 않아 불필요한 작업이 남을 수 있다. |
| P3 | 경미한 요청당 객체 생성 | `archway/gateway/adapter/in/middleware/request_middleware.go` | `json.NewEncoder`, bearer 문자열 정리, 에러 응답 생성이 요청마다 반복된다. 단독 효과는 작지만 핫패스라 누적된다. |

## 상세 메모

### 1. JWT 검증 핫패스 재사용성 부족
- 인증이 필요한 lookup에서 `getUserIdAndVerifyAccessToken`이 요청마다 `gjwt.NewCodec(keyService)`를 호출한다.
- 벤치마크 기준:
  - gRPC unix controller lookup: `HS256 26612 ns/op`, `RS256 57660 ns/op`
  - full-chain E2E: `NoJWT gRPC P8 14000 ns/op` → `RS256 gRPC P8 17941 ns/op`
- 수정 포인트:
  - 서비스별 codec 또는 검증기 재사용 구조 검토
  - route load 시점 초기화 가능 여부 확인
  - 인증 경로에서 불필요한 객체 생성 제거

### 2. 경로 파싱 로직 중복 + 문자열 분할/결합 반복
- 동일한 path 규칙 파싱이 gRPC handler, HTTP DTO, utils에 중복돼 있다.
- `strings.Split(strings.Trim(...), "/")` 와 `strings.Join(...)`가 반복되어 수정 누락 위험과 할당 비용이 동시에 존재한다.
- 수정 포인트:
  - 파서를 한 곳으로 통합
  - domain/path 추출 시 저할당 방식으로 정리
  - 동일 규칙을 `common/core` 수준에서 재사용

### 3. Trie 검색 시 요청당 map/문자열 재가공 발생
- `Search`가 매 요청마다 빈 `map[string]string`을 생성한다.
- 동적 path 매칭 후 `copyPathParams`와 `rewritePath`에서 다시 map 순회와 문자열 치환이 발생한다.
- 현재 구조는 path param이 실제로 필요 없는 요청에도 동일 비용을 지불한다.
- 수정 포인트:
  - path param 미사용 경로의 fast-path 분리
  - rewrite 결과 캐시 또는 세그먼트 기반 조합 검토
  - `Search` 반환 구조를 최소 데이터 중심으로 재정의

### 4. 캐시 초기화 파서의 전역 worker 구조
- `core/utils/upstream_parser.go`는 패키지 init 시 전역 worker 50개를 띄운다.
- 전역 채널 공유 구조라 호출 단위 격리가 약하고, startup scale이 커질수록 contention 추적이 어려워진다.
- `json.Unmarshal([]byte(request.payload), ...)`는 요청마다 문자열→바이트 슬라이스 변환 비용이 추가된다.
- 수정 포인트:
  - 호출 단위 worker pool 또는 bounded concurrency 전환
  - `GOMAXPROCS` 기반 동시성 조정
  - payload 표현 방식과 JSON decode 비용 재검토

### 5. gateway ↔ controller 호출에서 context 전파 부재
- controller gRPC lookup이 `context.Background()`로 호출된다.
- upstream 요청이 이미 취소됐어도 lookup RPC는 계속 진행될 수 있어 리소스 효율이 떨어진다.
- 수정 포인트:
  - HTTP 요청 context를 lookup 계층까지 전달
  - timeout/cancel을 controller RPC와 downstream transport에 반영

### 6. 경미한 요청당 객체 생성
- middleware에서 `json.NewEncoder(writer)`를 미리 만들고, authorization 문자열을 매번 정리한다.
- 큰 병목은 아니지만 요청량이 높아질수록 잔여 allocation을 줄일 수 있다.
- 수정 포인트:
  - 에러 응답 경로 전용 객체 생성으로 범위 축소
  - bearer token 정리/에러 응답 유틸 정리

## 추천 정리 순서
1. JWT 검증 재사용 구조
2. 경로 파싱 통합
3. Trie 검색 결과 구조 단순화
4. 캐시 초기화 파서 구조 개선
5. context 전파 보강
6. 소규모 allocation 정리

## 참고 근거
- `.docs/benchmark/readout-report.md`
- `.docs/benchmark/자료/gateway-controller-throughput-parallel-benchmark.txt`
- `.docs/benchmark/자료/gateway-fullchain-e2e-benchmark.txt`
