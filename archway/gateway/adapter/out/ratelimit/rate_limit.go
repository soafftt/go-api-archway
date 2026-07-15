package ratelimit

import (
	"gateway/application/port/out/ratelimiter"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var rwMutex = sync.RWMutex{}
var rateLimitCache = make(map[string]*rate.Limiter)

type rateLimit struct{}

func NewRateLimit() ratelimiter.RateLimiterPort {
	return &rateLimit{}
}

func (rl *rateLimit) Allow(service string, originalPath string, rateLimitCount int32) (bool, ratelimiter.RemainToken, ratelimiter.RemainTokensAt) {
	key := service + "|" + originalPath

	rwMutex.RLock()
	rateLimit, ok := rateLimitCache[key]
	rwMutex.RUnlock()
	if !ok {
		rwMutex.Lock()
		rateLimit, ok = rateLimitCache[key]
		if !ok {
			rateLimit = rate.NewLimiter(rate.Limit(rateLimitCount), int(rateLimitCount))
			rateLimitCache[key] = rateLimit
		}
		rwMutex.Unlock()
	}

	allowed := rateLimit.Allow()
	remainToken := ratelimiter.RemainToken(rateLimit.Tokens())
	remainTokenAt := ratelimiter.RemainTokensAt(rateLimit.TokensAt(time.Now().Add(1 * time.Second)))

	return allowed, remainToken, remainTokenAt

}
