# archway

`archway`는 두 개의 Go 애플리케이션으로 구성된 라우팅 게이트웨이 프로젝트입니다.

- `app/gateway`: 외부 요청을 수신하고 실제 업스트림으로 프록시하는 데이터 플레인
- `app/gateway-controller`: 라우팅 정책을 조회/제공하는 컨트롤 플레인

핵심 의도는 **요청 처리(프록시)와 라우팅 정책 결정(조회/관리)을 분리**하여, 빠른 요청 처리를 유지하면서 정책 변경에 유연하게 대응하는 것입니다.

## Why This Project

이 프로젝트는 다음 목적을 갖습니다.

- 요청 경로 기반 업스트림 라우팅
- 경로 재작성(path rewrite) 및 캐시 헤더 같은 게이트웨이 처리
- 라우팅 정책 로직을 별도 서비스로 분리하여 확장성 확보
- 공통 라우팅 모델/로직을 `common` 모듈로 공유

## Project Structure

```text
app/
  gateway/              # Reverse proxy 서버
  gateway-controller/   # Upstream 라우팅 조회 서버 (Unix socket)
common/
  model/                # 공통 도메인 모델 및 라우팅 로직
```

## Components

### 1) gateway

주요 역할:

- 클라이언트 HTTP 요청 수신
- 요청 경로를 기준으로 `gateway-controller`에 업스트림 조회
- 조회 결과를 기반으로 업스트림 URL 재작성 후 Reverse Proxy 수행
- 에러 응답/헤더 처리

특징:

- `UPSTREAM_LOOKUP_BASE_URL` 기본값: `http://unix/v1/upstream?path=`
- Unix socket 기반 HTTP 클라이언트를 통해 controller와 통신

### 2) gateway-controller

주요 역할:

- `/v1/upstream?path=...` 엔드포인트 제공
- 요청 path를 파싱해 서비스/도메인/리소스 경로 매칭
- 라우팅 규칙에 맞는 upstream 정보 반환

특징:

- Unix socket 서버로 동작 (`UNIX_SOCKET_PATH`)
- 라우팅 데이터 소스(예: Valkey)와 연동 가능한 구조

### 3) common

주요 역할:

- 업스트림 도메인 모델
- path router(Trie 기반) 등 재사용 가능한 라우팅 로직
- 서비스 간 공유되는 DTO/모델

## Request Flow

1. 클라이언트가 `gateway`로 요청을 보냄
2. `gateway`는 요청 path를 `gateway-controller`의 `/v1/upstream`으로 조회
3. `gateway-controller`가 path를 해석해 업스트림 정보 반환
4. `gateway`가 반환된 정보로 경로를 재작성하고 업스트림으로 프록시
5. 응답을 클라이언트에 전달

