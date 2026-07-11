package dto

type UpStreamLookupResult struct {
	Info  UpStreamInfo
	Error error
}

func NewUpStreamLookupResult(info UpStreamInfo) UpStreamLookupResult {
	return UpStreamLookupResult{
		Info: info,
	}
}

func NewErrUpStreamLookupResult(err error) UpStreamLookupResult {
	return UpStreamLookupResult{Error: err}
}

type UpStreamInfo struct {
	Host            string `json:"domain"`             // 도메인 이름
	Path            string `json:"path"`               // 프록시 경로
	Method          string `json:"method"`             // 메소드
	ResponseTimeout int64  `json:"response_timeout"`   // 응답 타임아웃
	RequestTimeout  int64  `json:"request_timeout"`    // 요청 타임아웃
	CacheTimeout    int64  `json:"cache_timeout"`      // 캐시 타임아웃
	UserKey         any    `json:"user_key:omitempty"` // 인증 체크이후의 유저키.
}

func NewUpStreamInfo(
	host, path, method string,
	requestTimeout, responseTimeout, cacheTimeout int64,
	userKey any,
) UpStreamInfo {
	return UpStreamInfo{
		Host:            host,
		Path:            path,
		Method:          method,
		ResponseTimeout: responseTimeout,
		RequestTimeout:  requestTimeout,
		CacheTimeout:    cacheTimeout,
		UserKey:         userKey,
	}
}
