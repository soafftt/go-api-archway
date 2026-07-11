package cache

import (
	cdu "core/domain/upstream"
)

type contextKey string

type RouteCache interface {
	LoadCache()
	Get(key string) (*cdu.UpstreamService, bool)
	Update(keys []string) error
	Evict(service string)
}
