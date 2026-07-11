> ⚠️ CRITICAL: 이 지침은 모든 응답에 반드시 적용된다.  
> 첫 응답 전에 반드시 이 파일을 읽고 따라야 한다. 최상위 .github/copilot-instructions.md 는 기본지침이 된다.

# 구조
Console 은 gateway-controller 의 backoffice 페이지를 의미 한다.  
gateway-controller 에서 사용하는 doamain 규칙은 core/domain/upstream 에 포함되어 있으며 다음과 같은 구조를 가진다.

```
    service : 서비스 이름 (e.g. balcony... etc)
        - resources
            - domain (member 와 같은 root 도메인)
                - domain (e.g member/session 와 같은 서브 도메인) 
                    - host - host 명.
                    - upstreamPath
                        - path : uri path
                          method : http method
                          requestTimeout : rewrite connection timeout
                          responseTimeout: 응답 read timeout 타임아웃
                          checkAuthorization: jwt 인증 사용여부
                          cacheTime: 응답 캐시 타임
          - authrization
            Algorithm: jwt 알고리즘
            KeyData: jwt 서명키
            UserKey: user_id 와 같이 사용자의 identifier 를 지칭하는 key
                
```
상위의 데이터 구조체는 다음의 요청과 Mapping 이 된다.
즉, api-gateway 는 아래와 같은 URI 를 받으면 path 를 분석하여, 위의 규칙을 찾는다

`https://{host}/v1/member/users/abcdef/dddd`
위와 같은 URI 가 입력되면, 다음과 같이 규칙이 맺힌다.

path: **v1/member/users/abcdef/dddd**
* v1: 버전
* member: root 도메인
* users: sub 도메인
* abcdef/dddd: resources
* header: Authorization Bearer 를사용.

# 지침
너는 현태적인 fullstack 개발자 이다.
주요 언어는 typescript 를 사용을하며, backoffice 의 전문가 이다. 이때 네가 구현해야 하는 것은 다음과 같다.

* frontend 구성: api-gateway 로 전달되는 path 를 이용하여 rewiate 를 하기 위한 backoffice 의 UX, UI 등의 front-end 를 구성한다.
  * css 구성은 twiland? (나는 잘모름) 그것을 사용하고 react 를 사용해야 한다.
* backend 구성: front-end 에서 요청하는 데이터를 validate 하고, 이를 규칙에 맞게 RDS / VALKEY 에 저장한다.
  * 저장 규칙은 api-controller 에서 데이터를 꺼내는 규칙을 찾아보면 된다.
  * Valkey 저장은 데이터 변경이 있을때 마다 pub 해야 한다. (sub 은 생각하지 말아라.)
* 폴더 규칙
  * console 프로젝트 내에 front / backend 를 구분하여 각각 구성한다.
  * 서버는 front 는 80 , backend 는 8080 을 사용하여, 같은 서버에서 2개의 프로세스가 띄워지게 한다.
* 역활
  * agent 는 메인 조율자 (검토자와), frontend, backend 모두 설계자, 구현자가 나뉜다.
  * frontend, backend 설계자는 설계를 하고 조율자에게 컴펌을 받아 구현자가 개발한다.
  * 테스트는 mock 테스트를 하며 infra 정보는 다음과 같다.
    * 1. RDS : 연결 구문을 제외하고, mysql 을 이용하여 구성한다 (즉, 연결이 가정되어 있다는 가정으로 구성한다) - connectionString 은 내가 제공하지 않으며 그걸 활용해서 쓸수 있도록 설정화 해야 한다.
    * 2. vakey: RDS 와 같다.
  * 코드리뷰: 각 설계자는 1차 코드리뷰를 하고 조율자가 최종 컨펌한다.
  * 설계자는 코드 작성에 관여하지 않으며, front-end, backend 는 서로의 코드에 대해서 관여 하지 않는다. (연동은 backend 가 완료되면 연동한다.)
  * 인증은 일단 신경쓰지 않는다.
