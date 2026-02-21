package proxy

/*
Proxy 메타 정보
*/
type RequestProxy struct {
	Path              string // path
	Method            string // allow method
	RequestTimeout    int64  // Request timeout
	ResponseTimeout   int64  // Response timeout
	CheckAuthrozation bool   // authorization check
}

type RequestService struct {
	ServiceName string         // benefit, member ... etc
	Host        string         // service host
	Request     []RequestProxy // proxy metadata
}

type RequestServiceCache struct {
	Services map[string]RequestService
}
