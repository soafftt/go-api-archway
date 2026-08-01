---
LOCKING: 
  - ../../agents/architecture.agent.md 는 설계를 시작하고 이후 지침대로 진행한다.
  - archway/gateway-controller/.github/agents/architecture.agent.mㅇ ,developer.agent.md 를 먼저 확인하고 작업 진행
---

#### 요구 사항
- grpc 구현 (Listener 포함) ../../../gatewat-controller/adapter/port/in/pb 에는 protobufer generate 파일이 있다.
- 여기까지만 작업이 되어 있고, 실제 서비스 구현 및 gprc 리스너를 구현하지 않았다.
- 기존 http unix_server 를 건들지 않고 main.go 에서 언제든 바꿔칠수 있는 grpc Listener 와 서비스 구현을 진행한다.
- grpc 의 서버 설정은 모범사례를 따르며, app_config.go 로 관리 할 수 있는 설정을 분리한다.
- 현재는 서비스가 한개 이지만, 더 늘어날 수 있으며 이를 고려하여, 모다 쉽게 구현된 서비스가 등록될 수 있도록 해야 한다.
  - e.g.  UpStreamRouter 참고.
- 개발이 완료 되면 developer.agent.md 에게 단위 테스트 및 benchmark 테스트를 요구 한다.
  - 테스트는 다음과 같다.
    - 1. 단순 서버서가 실행되는지.
    - 2. 기능 테스트가 통과 하는지.
    - 3. benchmark
      - nojwt 로의 성능테스트
      - HS256, RS256 성능테스트

