# 코드 리뷰 결과 (변경 사항 기준)

## 1. 오류 감지 및 예방 (필수)

### 🔴 **Critical Issue: Race Condition - 동시성 문제**
**위치**: `upstream_service.go` - `initializeResourceIndex()` 및 `LookupResourceDomain()`

**문제**:
```go
func (u *UpstreamService) initializeResourceIndex() {
    if u.resourceIndex != nil {  // ← 체크
        return
    }
    u.resourceIndex = make(map[string]*UpstreamResource, len(u.Resources))  // ← 이중 체크 락 패턴 없음
}
```
여러 고루틴에서 동시에 `LookupResourceDomain()` 호출 시:
- 두 고루틴이 동시에 nil 체크를 통과할 수 있음
- `u.resourceIndex`가 두 번 초기화될 수 있음 (데이터 레이스)
- Go의 메모리 보장이 깨짐

**영향**: 프로덕션에서 간헐적 데이터 손상/불일치 버그 발생

**권장 해결**: `sync.Once` initializeResourceIndex
```go
var once sync.Once
once.Do(u.initializeResourceIndex)
```

---

### 🟡 **High Priority: 오타 및 명명 규칙 위반**
**위치**: `upstream_path.go` - `CheckAuthrozation` 필드명

**문제**:
- 필드명: `CheckAuthrozation` (오타)
- 올바른 표기: `CheckAuthorization`
- JSON 태그: `check_authorization`
- 코드 일관성 저하, 유지보수 난제

**권장**: 모든 참조 위치 수정
- DTO: `CheckAuthorization`
- 기존 코드 호환성 주의 필요

---

### 🟡 **High Priority: JSON 태그 오타**
**위치**: `upstream_service.go` - `Resources` 필드

**문제**:
```go
Resources []*UpstreamResource `json:"resouces"`  // ← "resouces" 오타 (정상: "resources")
```
- 주석에 "matches API spec"이라고 명시되어 있으나
- API 규격도 함께 수정되어야 함 (아니면 의도적 별칭)
- **현재 상태**: 의도적 오타인 것으로 보이나, 명확한 문서화 필요

---

## 2. Go 철학 반영

### 🟡 **Unexported 함수에 주석 누락**
**위치**: `upstream_service.go` - `initializeResourceIndex()`

**문제**:
```go
func (u *UpstreamService) initializeResourceIndex() {  // ← 주석 없음
    // ...
}
```

**Effective Go 규칙**: 모든 exported/unexported 함수에 주석 필요

**수정**:
```go
// initializeResourceIndex initializes the internal sub-domain index map for fast lookups.
// This method is called lazily on first resource lookup.
func (u *UpstreamService) initializeResourceIndex() {
```

---

### 🟡 **Receiver 타입 일관성 없음**
**위치**: 전체 구조 검토

**문제**:
- `UpstreamService` 메서드: **포인터 리시버** (`*UpstreamService`)
- `UpstreamResource` 메서드: **포인터 리시버** (`*UpstreamResource`)
- 일관성은 있으나, **값 변경이 없는 메서드도 포인터 사용** (LookupPath는 읽기만 함)

**Effective Go 권장**:
- 값 수정 메서드 → 포인터 리시버 ✅
- 읽기 전용 메서드 → 값 리시버 권장

**예시**:
```go
// 현재 (포인터 리시버)
func (u *UpstreamResource) LookupPath(path string) *UpstreamPath {

// 권장 (값 리시버 - 변경 없음)
func (u UpstreamResource) LookupPath(path string) *UpstreamPath {
```

---

## 3. Effective Go 기반 개선사항

### 🟡 **Error Handling 부재**
**위치**: `upstream_service.go` - `LookupResourceDomain()`

**문제**:
```go
func (u *UpstreamService) LookupResourceDomain(subDomain string) (resource *UpstreamResource, isEmptyDomain bool)
```
- 에러 반환이 없음 (리소스 없을 때 nil만 반환)
- 호출자가 "없음"의 이유를 알 수 없음

**Effective Go 권장**:
```go
func (u *UpstreamService) LookupResourceDomain(subDomain string) (*UpstreamResource, bool, error) {
    // 리소스 없으면: return nil, false, fmt.Errorf("resource not found for subdomain: %s", subDomain)
}
```

---

### 🟡 **Type Assertion / Nil Check 미흡**
**위치**: `upstream_service.go` - `initializeResourceIndex()` 루프

**문제**:
```go
for _, resource := range u.Resources {
    if resource == nil {
        continue  // 조용히 skip - 로그 없음
    }
```

**Effective Go 권장**:
```go
for i, resource := range u.Resources {
    if resource == nil {
        // log.Printf("warning: nil resource at index %d", i)  // 디버깅 가능
        continue
    }
```

---

### 🟡 **Interface 미사용**
**문제**:
- `UpstreamService`, `UpstreamResource` 직접 사용
- 테스트 시 Mock 어려움, 결합도 높음

**Effective Go 권장**:
```go
type ResourceLookup interface {
    LookupResourceDomain(subDomain string) (*UpstreamResource, bool)
}
```

---

## 4. Go 성능 최적화

### 🟡 **Lazy Initialization의 동시성 문제**
**위치**: `upstream_service.go` + `upstream_domain.go`

**문제**:
```go
// upstream_domain.go - LookupPath
if u.pathRouter == nil {
    u.InitializeRouter()  // ← race condition 가능
}
```

**개선**: `sync.Once` 사용
```go
var once sync.Once
once.Do(u.InitializeRouter)
```

---

### 🟢 **좋은 점: Custom JSON Unmarshaler**
**위치**: `upstream_path.go` - `UnmarshalJSON`

**장점**:
- `cache-time` / `cache_timeout` 유연한 지원
- 성능 오버헤드 최소 (한 번만 실행)

---

## 5. 종합 개선 권고사항

| 우선순위 | 분류 | 내용 | 파일 |
|---------|------|-----|------|
| **P0** | Race Condition | `sync.Once` 도입하여 동시성 안전성 확보 | upstream_service.go, upstream_domain.go |
| **P1** | 명명 규칙 | `CheckAuthrozation` → `CheckAuthorization` 수정 | upstream_path.go + 전체 참조 |
| **P2** | 문서화 | unexported 함수 주석 추가 | upstream_service.go |
| **P2** | 에러 처리 | `LookupResourceDomain` error 반환 추가 | upstream_service.go |
| **P2** | Go 철학 | 읽기 전용 메서드는 값 리시버 사용 | upstream_domain.go |
| **P3** | 테스트성 | Interface 정의하여 Mock 지원 | upstream_service.go |

---

## 결론

✅ **현재 상태 요약**:
- 기능적으로는 정상 작동
- JSON 규격 정렬 우수
- 구조체 설계 명확

⚠️ **해결 필수**:
1. Race condition (`sync.Once` 도입)
2. 필드명 오타 수정
3. 주석 및 에러 처리 강화

🎯 **다음 단계**: P0/P1 항목 수정 후 재검토 권장
