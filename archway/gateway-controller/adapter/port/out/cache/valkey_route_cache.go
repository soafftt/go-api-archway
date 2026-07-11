package cache

import (
	"context"
	cdu "core/domain/upstream"
	"core/utils"
	"fmt"
	"gateway/controller/adapter/config"
	"strings"
	"sync"

	"github.com/valkey-io/valkey-go"
)

type RouteValkeyCache struct {
	client valkey.Client
	data   map[string]*cdu.UpstreamService
	mu     sync.RWMutex
}

func NewRouteValkeyCache(valkey config.ValkeyClient) *RouteValkeyCache {
	return &RouteValkeyCache{
		client: valkey.GetClient(),
		data:   make(map[string]*cdu.UpstreamService),
	}
}

func (r *RouteValkeyCache) LoadCache() {
	keys, err := r.keyScan(context.Background(), 0)
	if err != nil {
		panic(err)
	}
	if len(keys) == 0 {
		return
	}

	results, err := r.getValues(keys)
	if err != nil {
		panic(err)
	}
	for idx, _ := range results {
		result := results[idx]
		r.mu.Lock()
		r.data[result.ServiceName] = result
		r.mu.Unlock()
	}
}

func (r *RouteValkeyCache) Get(key string) (*cdu.UpstreamService, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, has := r.data[key]
	return data, has
}

func (r *RouteValkeyCache) Update(keys []string) error {
	results, err := r.getValues(buildValkeyKeys(keys))
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for idx := range results {
		r.data[results[idx].ServiceName] = results[idx]
	}
	return nil
}

func (r *RouteValkeyCache) Evict(service string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, service)
}

func (r *RouteValkeyCache) keyScan(
	ctx context.Context,
	cursor uint64,
) ([]string, error) {
	cmd := r.client.B().Scan().Cursor(cursor).Match("UPSTREAM:*").Build()
	result, err := r.client.Do(ctx, cmd).AsScanEntry()

	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	if result.Cursor > 0 {
		keys, err = r.keyScan(ctx, result.Cursor)
		if err != nil {
			return keys, err
		}
	}

	return append(result.Elements, keys...), nil
}

func (r *RouteValkeyCache) getValues(keys []string) ([]*cdu.UpstreamService, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	command := r.client.B().Mget().Key(keys...).Build()
	values, err := r.client.Do(context.Background(), command).AsStrSlice()
	if err != nil {
		return nil, fmt.Errorf("route cache mget: %w", err)
	}

	parseRequests := make([]utils.ParseRequest, len(values))
	for idx, _ := range keys {
		serviceName := strings.Replace(keys[idx], "UPSTREAM:", "", 1)
		parseRequests[idx] = utils.NewParseRequest(serviceName, values[idx])
	}

	return utils.ParseToUpstreamServiceWithInitialize(parseRequests), nil
}

func buildValkeyKeys(keys []string) []string {
	valkeyKeys := make([]string, len(keys))
	for idx := range keys {
		if strings.HasPrefix(keys[idx], "UPSTREAM:") {
			valkeyKeys[idx] = keys[idx]
			continue
		}
		valkeyKeys[idx] = "UPSTREAM:" + keys[idx]
	}
	return valkeyKeys
}
