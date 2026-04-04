package component

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	dto "gateway/common/model/upstream"
	"gateway/controller/infra"
)

// 서비스별 업스트림 라우팅 정보 저장 키 패턴
var upstreamKeyPattern = "UPSTREAM:*"

type RouteCache interface {
	Get(service string) (*dto.UpstreamService, bool)
	Update(ctx context.Context, keys []string) error
	Evict(service string)
}

type routeCache struct {
	valkey *infra.ValkeyWrap
	parser RouteParse
	data   map[string]*dto.UpstreamService
}

func NewUpstreamRouteCache(valkey *infra.ValkeyWrap, parser RouteParse) *routeCache {
	cache := &routeCache{
		valkey: valkey,
		parser: parser,
		data:   make(map[string]*dto.UpstreamService),
	}
	cache.initializeCache()

	return cache
}

func (u *routeCache) initializeCache() {
	ctx := context.Background()
	keys, err := u.getUpstreamKeys(ctx, 0, nil)
	if err != nil {
		log.Fatalf("initialize upstream cache error: %v", err)
	}

	u.Update(ctx, keys)
}

/*
valkey 에서 "UPSTREAM:*" 패턴의 키를 모두 조회
*/
func (u *routeCache) getUpstreamKeys(
	ctx context.Context,
	cursor uint64,
	beforeKeys []string,
) ([]string, error) {
	valkeyClient := u.valkey.GetClient()
	keyCount := int64(1000)

	result := valkeyClient.Do(
		ctx,
		valkeyClient.B().Scan().Cursor(cursor).Match(upstreamKeyPattern).Count(keyCount).Build(),
	)

	scanEntity, err := result.AsScanEntry()
	if err != nil {
		return nil, fmt.Errorf("Error scanning keys: %v", result.Error())
	}

	if beforeKeys == nil {
		beforeKeys = make([]string, 0)
	}

	keyResults := append(beforeKeys, scanEntity.Elements...)
	if scanEntity.Cursor != 0 {
		// 재귀 호출로 다음 페이지의 키 조회
		return u.getUpstreamKeys(ctx, scanEntity.Cursor, keyResults)
	}

	return keyResults, nil
}

func (u *routeCache) Get(service string) (*dto.UpstreamService, bool) {
	k, ok := u.data[service]
	return k, ok
}

/*
조회된 키에 대해 MGET 명령으로 일괄 조회하여 정책 데이터를 가져옴
- 가져온 정책 데이터를 RouteParser를 통해 UpstreamService 구조체로 파싱하여 반환
- 파싱된 UpstreamService 구조체를 data 맵에 저장하여 서비스/호스트/URI별 정책을 관리
*/
func (u *routeCache) Update(ctx context.Context, keys []string) error {
	valkeyClient := u.valkey.GetClient()

	mgetCommand := valkeyClient.B().Mget().Key(keys...).Cache()
	mgetResult := valkeyClient.DoCache(
		ctx,
		mgetCommand,
		10*time.Minute,
	)

	values, err := mgetResult.AsStrSlice()
	if err != nil {
		return fmt.Errorf("Error fetching values for keys: %v", err)
	}

	sources := make([]RouteSource, len(keys))
	for index, key := range keys {
		serviceName := strings.Replace(
			key,
			"UPSTREAM:",
			"",
			1,
		)

		if values[index] == "" {
			// 키는 존재하지만 값이 없는 경우, 즉 정책이 삭제된 경우 캐시에서 제거
			u.Evict(serviceName)
		} else {
			fmt.Println(values[index])
			sources[index] = NewRouteSource(serviceName, values[index])
		}
	}

	upstreams, err := u.parser.ParseFromSlice(sources)
	if err != nil {
		return fmt.Errorf("Error parsing upstreams: %v", err)
	}

	for _, data := range upstreams {
		u.data[data.ServiceName] = data
	}

	return nil
}

func (u *routeCache) Evict(service string) {
	delete(u.data, service)
}
