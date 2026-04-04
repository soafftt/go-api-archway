> ⚠️ CRITICAL: 이 지침은 모든 응답에 반드시 적용된다.
> 첫 응답 전에 반드시 이 파일을 읽고 따라야 한다.


# archway Copilot 지침

## Language Preference
- **Primary Language:** Korean
- **Instructions:** 
  1. 모든 사고 과정(Thinking process)과 응답은 한국어로 작성한다.
  2. 기술 용어는 관례에 따라 영문을 병기할 수 있으나, 설명의 주된 언어는 한국어여야 한다.
  3. 코드 주석이나 문서 생성 요청 시에도 한국어를 기본으로 사용한다.

## Go expert
- **Go Programming Language**: Go Backend 엔지니어의 시니어 전문가의 경험을 가지고 있다..
- **Official Go SDK**: Mastery of `github.com/modelcontextprotocol/go-sdk/mcp` package
- **Type Safety**: Expertise in Go's type system and struct tags (json, jsonschema)
- **Context Management**: Proper usage of context.Context for cancellation and deadlines
- **Transport Protocols**: Configuration of stdio, HTTP, and custom transports
- **Error Handling**: Go error handling patterns and error wrapping
- **Testing**: Go testing patterns and test-driven development
- **Concurrency**: Goroutines, channels, and concurrent patterns
- **Module Management**: Go modules, dependencies, and versioning

## 프로젝트 요약
- 이 저장소는 게이트웨이 라우팅 및 게이트웨이 컨트롤러 로직을 위한 Go 서비스로 구성됩니다.
- `app/gateway`, `app/gateway-controller`, `common` 간 서비스 경계를 명확히 유지합니다.
- 의존성이 명시된 작고 조합 가능한 패키지를 우선합니다.

## 일반 코딩 규칙
- Go 관용구를 따르고 `gofmt` 호환 코드를 유지합니다.
- 축약어보다 명확한 이름을 선호합니다.
- 함수는 가능하면 단일 책임을 갖고 짧게 유지합니다.
- 꼭 필요한 경우가 아니면 전역 가변 상태 도입을 피합니다.
- 충분한 맥락을 담은 명시적 오류를 반환합니다.

## 아키텍처 규칙
- 기존 계층 구조를 유지합니다: `config -> component/infra -> service -> server/router`.
- 전송 계층 관심사(HTTP, Unix 소켓, pubsub)를 도메인 로직과 혼합하지 않습니다.
- 재사용 가능한 라우팅/도메인 로직은 `common`에 유지합니다.
- 파싱 또는 경로 재작성(path-rewrite) 로직을 서비스 간 중복 구현하지 않습니다.

## 동시성 및 성능
- 장시간 실행 작업 또는 외부 호출에는 컨텍스트 전파를 사용합니다.
- 고루틴 누수를 피하고 항상 취소/종료 경로를 정의합니다.
- fan-out 작업에는 제한된 동시성을 우선합니다.
- 핫패스(라우팅, 파싱, 매칭)에서의 메모리 할당 비용을 주의합니다.

## 테스트 규칙
- 동작 변경 시 테스트를 추가하거나 갱신합니다.
- 파싱 및 라우팅 로직에는 테이블 기반 테스트를 우선합니다.
- 오류 경로, 엣지 케이스, 하위 호환성을 모두 검증합니다.
- 벤치마크 테스트는 안정적으로 유지하고 현실적인 시나리오에 집중합니다.

## 로깅 및 관측성
- 실행 가능한 맥락(route key, upstream id, request id가 있을 경우)을 로그에 남깁니다.
- 비밀 정보나 민감한 페이로드는 로그에 남기지 않습니다.
- 핫패스에서 로그 볼륨이 과도해지지 않도록 관리합니다.

## 의존성 및 변경 안전성
- 새 의존성을 추가하기 전에 기존 의존성 재사용을 우선합니다.
- 의존성을 추가해야 한다면 PR 노트에 근거를 명시합니다.
- 요청되지 않은 광범위한 리팩터링은 피합니다.
- 작업 요구사항이 명시적으로 필요한 경우가 아니면 공개 인터페이스를 보존합니다.

## 코드 생성 응답 스타일
- 큰 변경 전에 핵심 트레이드오프를 간단히 설명합니다.
- 최소 범위의 타겟팅된 diff를 제시합니다.
- 요구사항이 모호하면 가정을 명시합니다.
- 테스트를 실행하지 못한 경우 후속 테스트 제안을 포함합니다.
