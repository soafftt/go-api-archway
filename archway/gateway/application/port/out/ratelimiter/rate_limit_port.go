package ratelimiter

type RemainToken float64
type RemainTokensAt uint64

type RateLimiterPort interface {
	Allow(service string, originalPath string, rateLimitCount int32) (bool, RemainToken, RemainTokensAt)
}
