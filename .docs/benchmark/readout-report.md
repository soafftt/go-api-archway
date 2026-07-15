# Archway API Gateway - 100k RPS 기준의 관점

## 1. 목적
- 본 문서는 사내 자체개발 Archway API Gateway - 100k RPS 기준의 관점 용량/성능 기준 보고서입니다.
- 범위는 `RateLimit + JWT 검증`이 포함된 게이트웨이 요청 경로입니다.

## 2. 요약
- 결론: **조건부 내부 운영 가능** (Pod당 RateLimit의 정합성이 100% 보장되지 않음)
  - 동기화를 위한 Valkey/Readis 사용등을 고려할 수 있지만 이는 네트워크 비용상승에 대한 trade-off 임.
  - **단, 추후 Valkey/Redis 사용도 한번 고려해보고 benchmark 가 필요함**
- 권장 운영안:
  1. 인증 알고리즘 기준으로 용량 계획을 분리(HS256 전용 / RS256 전용 / 혼합)
  2. 목표 처리량 **100k RPS** 기준으로 Pod/Node를 N+1 여유로 배치
  3. 운영 전 검증 게이트(5xx, P99, CPU/RSS, 스케일 안정성) 통과

## 3. 용어
- `ms/op`: 요청 1건 평균 처리 시간
- `rps_p50/p95/p99`: 초당 처리량 분포 지표
  - P50: 평균 50%의 지표
  - P95: 평균 95%의 지표
  - P99: 평규 99%의 지표
- E2E : End to End 
- 상세 수치는 `.docs/benchmark/rate-limit-benchmark-table-full.md` 참조

## 4. 핵심 성능 결과 (E2E 기준)
| 시나리오 | HS256 (ms/op) | RS256 (ms/op) | 해석 |
|---|---:|---:|---|
| URI 100 (모든 URI에 RateLimit 적용) | 0.007861 | 0.010261 | RS256이 약 30% 느림 |
| URI 1000 (모든 URI에 RateLimit 적용) | 0.007767 | 0.009964 | URI 수 증가는 주 병목 아님 |
| URI 100~1000 (RateLimit 분포 변화) | 0.007697~0.007892 | 0.009964~0.010328 | RateLimit 분포 영향은 상대적으로 작음 |

## 5. 100k RPS 기준 권장 Pod 자원 (scale-out)
> 전제: RR 분산, 현재 구조 유지, 운영 여유 30~40%

| 인증 프로파일 | 권장 Pod 수 | Pod당 목표 RPS | Pod CPU (request/limit) | Pod RSS Memory 목표 |
|---|---:|---:|---:|---:|
| HS256 전용 | 6~8 | 12,500~16,667 | 2 / 3 vCPU | 450~700 MiB |
| 혼합(HS/RS=50:50) | 8~10 | 10,000~12,500 | 2.5 / 3.5 vCPU | 550~850 MiB |
| RS256 전용 | 9~12 | 8,333~11,111 | 3 / 4.5 vCPU | 650~1000 MiB |

## 6. Cluster Node 용량 예상 (여유분 포함)
기준 노드: `8 vCPU / 32 GiB`

### 6.1 산정식
1. 총 CPU request = Pod 수 × Pod CPU request  
2. 총 Memory request = Pod 수 × Pod Memory request  
3. 운영 여유 반영 = 총량 × 1.3

### 6.2 프로파일별 총량
| 인증 프로파일 | 총 CPU request | 총 Memory request | 여유(30%) 반영 CPU | 여유(30%) 반영 Memory |
|---|---:|---:|---:|---:|
| HS256 전용 (8 Pods) | 16 vCPU | 3.6~5.6 GiB | 20.8 vCPU | 4.7~7.3 GiB |
| 혼합 50:50 (10 Pods) | 25 vCPU | 5.5~8.5 GiB | 32.5 vCPU | 7.2~11.1 GiB |
| RS256 전용 (12 Pods) | 36 vCPU | 7.8~12.0 GiB | 46.8 vCPU | 10.1~15.6 GiB |

### 6.3 권장 노드 수
| 운영 모드 | 권장 워커 노드 |
|---|---:|
| HS256 전용 | 4~5 nodes |
| 혼합(50:50) | 5 nodes |
| RS256 전용 | 7 nodes |

## 7. 내부 운영 판정 기준
### 7.1 Go
- **100k RPS** 부하 30분에서 모두 충족:
  - Gateway 5xx < 0.1%
  - P99 지연 SLO 충족
  - Pod CPU/RSS 70% 이하
  - HPA scale-out 시 처리량/지연 악화 없음

### 7.2 Conditional Go
- 일부 미달이나 아래 보완안 확정 시:
  - RS256 비중 조정 또는 상위 인증 계층 분리
  - Pod 1~2개 증설
  - Node 1개 증설

## 8. 전제 및 리스크
1. 본 수치는 단일 테스트 환경 기준이며, 운영 절대값은 달라질 수 있음  
2. 현재 RateLimit은 pod-local 메모리 기반이므로 RR 분산 시 정책 정합성 리스크 존재  
3. 본 결론은 “정합성 리스크를 인지/수용한다”는 전제에서의 용량 관점 판정

## 9. 참조 문서
- `.docs/benchmark/gateway-middleware-benchmark.txt`
- `.docs/benchmark/gateway-middleware.cpu.pprof`
- `.docs/benchmark/gateway-middleware.mem.pprof`
- `.docs/benchmark/rate-limit-benchmark-table-full.md`
