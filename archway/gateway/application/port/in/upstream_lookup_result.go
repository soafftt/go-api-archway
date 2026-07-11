package in

import "strings"

// Lookup 결과.
type UpstreamLookupResult struct {
	Host            string // 도메인 이름
	Path            string // 프록시 경로
	Method          string // 메소드
	ResponseTimeout int64  // 응답 타임아웃
	RequestTimeout  int64  // 요청 타임아웃
	CacheTimeout    int64  // 캐시 타임아웃
	UserKey         any    // 인증 체크이후의 유저키.
}

// udp 나 tcp 를 지원하지 않기에 http/s 만 ....
func (u UpstreamLookupResult) Scheme() string {
	if ok := strings.Contains(strings.ToLower(u.Host), "https://"); ok {
		return "https"
	}

	return "http"
}

func (u UpstreamLookupResult) GetDomain() string {
	scheme := u.Scheme()
	return strings.ReplaceAll(u.Method, scheme+"://", "")
}

func (u UpstreamLookupResult) FullUrl() string {
	return u.Host + "/" + u.Path
}
