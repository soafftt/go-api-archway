# gateway-controller gRPC Benchmark Report

## 1. 목적
- `gateway-controller`의 upstream lookup 경로에서 **gRPC vs Unix Socket HTTP** 성능을 비교한다.
- `UserKey string` 적용, DTO 제거(파라미터 전달), unix socket 버퍼 조정 이후 결과를 보정/재정리한다.

## 2. 측정 환경/조건
- OS/CPU: `darwin arm64`, `Apple M4`
- 패키지: `gateway/controller/adapter/config`
- JWT 키/codec: 서비스 기동 시 캐시되어 재사용
- Valkey: `VALKEY_MASTER_HOST=127.0.0.1:6379`

## 3. 실행 커맨드

### 3.1 gRPC(bufconn) + gRPC(기본 벤치)
```bash
go test ./adapter/config -run '^$' -bench 'BenchmarkGrpcLookup_(NoJWT|HS256|RS256)' -benchmem
```

### 3.2 Unix Socket HTTP (Valkey/JWT)
```bash
VALKEY_MASTER_HOST=127.0.0.1:6379 \
go test ./adapter/config -run '^$' \
  -bench 'BenchmarkUnixServerLookupRouteOverUnixSocketWithValkeyAndJWT/(NoJWT|HS256|RS256)$' \
  -benchmem
```

### 3.3 조건 통일: gRPC도 Unix Socket Listener로 측정
```bash
VALKEY_MASTER_HOST=127.0.0.1:6379 \
go test ./adapter/config -run '^$' \
  -bench 'Benchmark(GrpcLookupOverUnixSocketWithValkeyAndJWT|UnixServerLookupRouteOverUnixSocketWithValkeyAndJWT)/(NoJWT|HS256|RS256)$' \
  -benchmem
```

## 4. 구간별 결과

### 4.1 최신 기준: gRPC(bufconn) vs Unix Socket HTTP
| Case | gRPC | Unix Socket HTTP |
|---|---:|---:|
| NoJWT | 15829 ns/op, 9833 B/op, 166 allocs/op | 19513 ns/op, 7097 B/op, 82 allocs/op |
| HS256 | 23729 ns/op, 13552 B/op, 220 allocs/op | 24526 ns/op, 11201 B/op, 140 allocs/op |
| RS256 | 49364 ns/op, 15797 B/op, 225 allocs/op | 51858 ns/op, 12748 B/op, 145 allocs/op |

관찰:
- gRPC(bufconn)는 `ns/op`가 유리하나 `B/op`, `allocs/op`는 HTTP보다 높다.

### 4.2 조건 통일: gRPC(unix socket) vs HTTP(unix socket)
| Case | gRPC over Unix Socket | HTTP over Unix Socket |
|---|---:|---:|
| NoJWT | 24562 ns/op, 9995 B/op, 166 allocs/op | 19887 ns/op, 7121 B/op, 82 allocs/op |
| HS256 | 33411 ns/op, 13834 B/op, 221 allocs/op | 24387 ns/op, 11198 B/op, 140 allocs/op |
| RS256 | 61734 ns/op, 16037 B/op, 225 allocs/op | 52405 ns/op, 12780 B/op, 145 allocs/op |

관찰:
- transport를 통일하면 gRPC가 시간/메모리/할당 모두 불리하다.
- 즉, `bufconn` 수치는 실제 unix socket 경로와 분리해서 해석해야 한다.

### 4.3 Unix Socket 버퍼(읽기/쓰기 1MiB) 증대 후
| Case | gRPC over Unix Socket (1MiB) | HTTP over Unix Socket (1MiB) |
|---|---:|---:|
| NoJWT | 24757 ns/op, 9935 B/op, 166 allocs/op | 19725 ns/op, 7090 B/op, 82 allocs/op |
| HS256 | 33688 ns/op, 13813 B/op, 221 allocs/op | 24463 ns/op, 11171 B/op, 140 allocs/op |
| RS256 | 61588 ns/op, 16066 B/op, 225 allocs/op | 51832 ns/op, 12773 B/op, 145 allocs/op |

관찰:
- 버퍼 증대 효과는 노이즈 수준(소폭 변동)이며 alloc은 사실상 동일.

### 4.4 DTO 제거(파라미터 전달) 전/후 비교 (gRPC bufconn)
| Case | DTO 방식(이전) | 파라미터 방식(현재) | 변화 |
|---|---:|---:|---:|
| NoJWT | 16226 ns/op, 9838 B/op, 166 allocs/op | 15829 ns/op, 9833 B/op, 166 allocs/op | -397 ns/op, -5 B/op, alloc 동일 |
| HS256 | 23742 ns/op, 13565 B/op, 220 allocs/op | 23729 ns/op, 13552 B/op, 220 allocs/op | -13 ns/op, -13 B/op, alloc 동일 |
| RS256 | 49625 ns/op, 15807 B/op, 225 allocs/op | 49364 ns/op, 15797 B/op, 225 allocs/op | -261 ns/op, -10 B/op, alloc 동일 |

### 4.5 JWT 캐시 재사용 검증 결과
- 실행:
  - `go test ./core/gjwt -run 'Test(RegisterECDSAKey|RegisterRSAKey|HMACCodecRoundTrip|ECDSACodecParse|RSACodecParse)'`
  - `go test ./core/gjwt -run '^$' -bench 'Benchmark(ECDSARegisterKey|HMACDeserialize|RSADeserialize)$' -benchmem`
- 결과:
  - 테스트 PASS
  - `BenchmarkECDSARegisterKey`: `6.558 ns/op`, `0 B/op`, `0 allocs/op` (이미 등록된 키 fast-path 무할당)
  - `BenchmarkHMACDeserialize`: `1592 ns/op`, `2392 B/op`, `48 allocs/op`
  - `BenchmarkRSADeserialize`: `23174 ns/op`, `3640 B/op`, `53 allocs/op`

### 4.6 gRPC Listener 비교: Unix Socket vs TCP (동일 벤치/동일 조건)
실행:
```bash
VALKEY_MASTER_HOST=127.0.0.1:6379 \
go test ./adapter/config/server/grpc_server -run '^$' \
  -bench 'BenchmarkGrpcLookupOver(UnixSocket|TCP)WithValkeyAndJWT/(NoJWT|HS256|RS256)$' \
  -benchmem
```

결과:
| Case | gRPC Unix Listener | gRPC TCP Listener |
|---|---:|---:|
| NoJWT | 23545 ns/op, 9998 B/op, 166 allocs/op | 33931 ns/op, 9929 B/op, 165 allocs/op |
| HS256 | 32006 ns/op, 13813 B/op, 221 allocs/op | 43441 ns/op, 13739 B/op, 221 allocs/op |
| RS256 | 58905 ns/op, 16025 B/op, 225 allocs/op | 70924 ns/op, 16213 B/op, 226 allocs/op |

관찰:
- 로컬 환경에서 gRPC listener를 unix socket으로 둘 때 TCP(127.0.0.1)보다 `ns/op`가 일관되게 낮았다.
- `B/op`, `allocs/op`는 거의 유사하며 주요 차이는 지연시간이다.

## 5. gRPC 종특(특성) 설명
- gRPC는 unary 호출마다 프레임워크 레벨 처리(stream/metadata/status, protobuf 메시지 처리)가 들어가므로, `allocs/op`, `B/op`가 HTTP보다 높게 나오는 패턴이 흔하다.
- 반대로 `bufconn` 같은 인메모리 경로에서는 커널/소켓 경로 비용이 빠져 `ns/op`가 유리하게 나올 수 있다.
- 따라서 gRPC 성능 평가는 `bufconn`과 실제 `unix socket` 결과를 분리해 해석해야 한다.
- `sync.Pool`로 요청 객체를 재사용하는 방식은 이 경로에서 효과가 제한적일 수 있다.
  - proto message는 내부 상태/참조 필드가 있어 재사용 시 초기화 누락 리스크가 크며, 잘못 쓰면 데이터 오염/경합 버그로 이어질 수 있다.
  - 현재 병목은 소켓 버퍼보다 상위 계층의 프레임워크/직렬화 오버헤드 비중이 커서, 객체 풀링만으로 큰 폭 개선이 나오기 어렵다.
  - 적용 시에는 proto 본체보다 보조 버퍼/임시 객체 범위에 제한하는 것이 안전하다.

## 6. 결론
- 현재 전략은 **gRPC latency 우선, 메모리 alloc 일부 수용**으로 정리한다.
- `UserKey string`, DTO 제거는 소폭 개선은 있었지만 구조적 alloc 차이를 뒤집지는 못했다.
- 병목/오버헤드는 소켓 버퍼보다 상위 계층(gRPC 프레임워크 처리, 직렬화, 요청당 객체 생성)에 더 가깝다.
- 로컬 IPC 전제에서는 gRPC listener를 TCP보다 unix socket으로 두는 편이 성능상 유리하다.
