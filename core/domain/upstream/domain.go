package upstream

type UpstreamResource struct {
	SubDomain  string              `json:"sub_domain"` // gateway 규칙의 서브도메인 (없으면 빈 문자열)
	Host       string              `json:"host"`       // service host
	Paths      []*UpstreamPath     `json:"paths"`      // proxy metadata
	pathRouter *UpStreamPathRouter // internal Trie router for fast path lookup
}

// InitializeRouter 는 path 정보를 기반으로 Trie 라우터를 초기화합니다. UpstreamService 의 InitializeResourceIndex에서 호출됩니다.
func (u *UpstreamResource) InitializeRouter() {
	u.pathRouter = NewUpStreamPathRouter()
	for _, p := range u.Paths {
		u.pathRouter.Insert(p)
	}
}

// LookupPath 는 Trie 기반 라우팅을 사용하여 일치하는 경로를 찾습니다.
// 정확한 일치와 경로 변수(예: /users/{id})를 모두 지원합니다.
func (u *UpstreamResource) LookupPath(path string) *UpstreamPath {
	if u.pathRouter == nil {
		return nil
	}

	return u.pathRouter.Search(path)
}
