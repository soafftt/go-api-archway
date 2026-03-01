package upstream

type UpstreamPath struct {
	Path               string `json:"path"`                // proxy path
	Method             string `json:"method"`              // allow method
	RequestTimeout     int64  `json:"request_timeout"`     // Request timeout
	ResponseTimeout    int64  `json:"response_timeout"`    // Response timeout
	CheckAuthorization bool   `json:"check_authorization"` // authorization check
	CacheTimeout       int64  `json:"cache_timeout"`       // Cache timeout
}
