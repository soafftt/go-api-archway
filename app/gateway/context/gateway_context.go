package context

type ReverseProxyContextKey string

const UpstreamContextKey ReverseProxyContextKey = "upstream"

const UserSessionContextKey ReverseProxyContextKey = "userSession"
