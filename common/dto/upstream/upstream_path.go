package upstream

type UpstreamPath struct {
	Path              string `json:"path"`               // proxy path
	Method            string `json:"method"`             // allow method
	RequestTimeout    int64  `json:"requestTimeout"`     // Request timeout
	ResponseTimeout   int64  `json:"responseTimeout"`    // Response timeout
	CheckAuthrozation bool   `json:"checkAuthorization"` // authorization check
}
