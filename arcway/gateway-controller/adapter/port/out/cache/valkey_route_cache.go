package cache

import (
	"context"
	cdu "core/domain/upstream"
	"core/utils"
	"gateway/controller/adapter/config"
	"log"
	"strings"

	"github.com/valkey-io/valkey-go"
)

type RouteValkeyCache struct {
	client valkey.Client
	data   map[string]*cdu.UpstreamService
}

func NewRouteValkeyCache(client *config.ValkeyClient) *RouteValkeyCache {
	return &RouteValkeyCache{
		client: client.SingleClient,
		data:   make(map[string]*cdu.UpstreamService),
	}
}

func (r RouteValkeyCache) LoadCache() {
	keys, err := r.keyScan(context.Background(), 0)
	if err != nil {
		panic(err)
	}
	if len(keys) == 0 {
		return
	}

	results := r.getValues(keys)
	for idx, _ := range results {
		result := results[idx]
		r.data[result.ServiceName] = result
	}
}

func (r RouteValkeyCache) Get(key string) (*cdu.UpstreamService, bool) {
	data, has := r.data[key]
	return data, has
}

func (r RouteValkeyCache) Update(keys []string) error {
	panic("implement me")
}

func (r RouteValkeyCache) Evict(service string) {
	delete(r.data, service)
}

func (r RouteValkeyCache) keyScan(
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
		keys, err := r.keyScan(ctx, result.Cursor)
		if err != nil {
			return keys, err
		}
	}

	return append(result.Elements, keys...), nil
}

func (r RouteValkeyCache) getValues(keys []string) []*cdu.UpstreamService {
	if len(keys) == 0 {
		return nil
	}

	command := r.client.B().Mget().Key(keys...).Build()
	values, err := r.client.Do(context.Background(), command).AsStrSlice()
	if err != nil {
		log.Fatal(err)
	}

	parseRequests := make([]utils.ParseRequest, len(values))
	for idx, _ := range keys {
		serviceName := strings.Replace(keys[idx], "UPSTREAM:", "", 1)
		parseRequests[idx] = utils.NewParseRequest(serviceName, values[idx])
	}

	return utils.ParseToUpstreamServiceWithInitialize(parseRequests)
}
