package cache

import cdp "core/domain/upstream"

type RouteCacheDto struct {
	data map[string]*cdp.UpstreamService
}

func NewRouteCacheDto() *RouteCacheDto {
	return &RouteCacheDto{
		data: make(map[string]*cdp.UpstreamService),
	}
}

func (r *RouteCacheDto) Get(key string) (*cdp.UpstreamService, bool) {
	service, has := r.data[key]
	return service, has
}

func (r *RouteCacheDto) Set(key string, value *cdp.UpstreamService) {
	r.data[key] = value
}

func (r *RouteCacheDto) Delete(key string) {
	delete(r.data, key)
}
