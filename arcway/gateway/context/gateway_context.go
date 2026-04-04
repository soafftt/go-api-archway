package context

type ReverseProxyContextKey string

const (
	UpstreamContextKey    ReverseProxyContextKey = "upstream"
	UserSessionContextKey ReverseProxyContextKey = "userSession"
)
