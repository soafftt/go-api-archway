# Gateway RateLimit 벤치마크 보고서

## 개요
- 대상: `archway/gateway/adapter/in/middleware/request_middleware.go`
- 기준 키: `Service:OriginPath`
- 목적: URI 수/RateLimit 분포/JWT 알고리즘(HS256, RS256) 조합에서 성능 및 E2E 영향 확인

## 시나리오
- URI 수: `100`, `1000`
- RateLimit 분포:
  - `all`: 모든 URI에 상이한 RateLimit
  - `2:3`: RateLimit 활성 40%, 비활성 60%
  - `1:3`: RateLimit 활성 25%, 비활성 75%
- JWT: URI 기준 50:50으로 체크/미체크 분배
- 알고리즘: `HS256`, `RS256`
- 지표: `rps_p50/p95/p99`, `ns/op`, `B/op`, `allocs/op`, `cpu.pprof`, `mem.pprof`

## 실행 커맨드
```bash
cd archway/gateway
go test ./adapter/in/middleware \
  -run '^$' \
  -bench 'Benchmark(RequestMiddleware_RateLimitMatrix|GatewayPipeline_RateLimitMatrixE2E)$' \
  -benchmem \
  -benchtime=1s \
  -count=1 \
  -cpuprofile ../.docs/benchmark/gateway-middleware.cpu.pprof \
  -memprofile ../.docs/benchmark/gateway-middleware.mem.pprof \
  | tee ../.docs/benchmark/gateway-middleware-benchmark.txt
```

## 핵심 결과 (middleware 단독)
1. URI100 All/P1: HS `889.2ns`, RS `3346ns` (RS 약 `3.76x` 느림)
2. URI1000 All/P1: HS `690.0ns`, RS `3332ns` (RS 약 `4.83x` 느림)
3. HS는 URI100→1000에서 All/P1 기준 `889.2ns → 690.0ns` (약 `22.4%` 개선)
4. RS는 URI100→1000에서 All/P1 기준 `3346ns → 3332ns` (거의 동일)
5. HS(URI100/P1)는 RL 분포가 희박할수록 개선: All `889.2ns`, 2:3 `739.0ns`, 1:3 `724.3ns`
6. RS(URI100/P1)는 RateLimit 분포 영향이 작음: All `3346ns`, 2:3 `3431ns`, 1:3 `3461ns`

## E2E 영향 (GatewayPipeline)
1. URI100 All: HS `7861ns`, RS `10261ns` / alloc `116~118`
2. URI1000 All: HS `7767ns`, RS `9964ns` / alloc `116~118`
3. middleware 단독 대비 E2E에서 지연과 할당이 크게 증가 (`~9.7KB~11.1KB`, `115~118 allocs/op`)

## 결론
- JWT 검증이 포함된 경로에서 `RS256` 비용이 `HS256` 대비 유의미하게 높다.
- RateLimit 미적용 URI 비율이 높아질수록(특히 HS) 미들웨어 처리 비용이 감소한다.
- 실제 서비스 체감 성능은 middleware 단독보다 E2E 경로 오버헤드의 영향을 더 크게 받는다.

## 에이전트 교차검증 결과
- 1) 시나리오/조율자: 요구사항 매핑 PASS, RL 비율 해석 문구 보강 권고
- 2) backend 개발: 구현/아티팩트 확인 PASS, 응답 헤더 순서/오탈자 리스크 지적
- 3) 분석자: 수치 요약 및 E2E 영향 정리 완료

## 산출물 경로
- `.docs/benchmark/gateway-middleware-benchmark.txt`
- `.docs/benchmark/gateway-middleware.cpu.pprof`
- `.docs/benchmark/gateway-middleware.mem.pprof`
- `.docs/benchmark/rate-limit-benchmark-table-full.md`
- `.docs/benchmark/readout-report.md`

## 주의사항
1. 본 결과는 단일 환경(darwin/arm64, Apple M4) 측정치다.
2. `rps_p50/p95/p99`는 1초 윈도우 샘플 기반이며, 운영 트래픽과 1:1 대응하지 않는다.
3. 비교 시 동일 parallel 수준(P값)끼리만 해석해야 한다.
