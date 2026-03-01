package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/controller/infra"
	"log"
	"strings"
	"time"

	dto "gateway/common/dto/upstream"
	modelDto "gateway/controller/model/dto"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type PolicyService interface {
	CheckPolicy(urlParseDto modelDto.URLParseDTO) (modelDto.RewitePathDTO, error)
}

type policyService struct {
	valkeyClient *infra.GlideValkey
}

// 서비스 / host / uri 의 정책 (allow or deny) 을 관리하는 서비스
var policyMap map[string]*dto.UpstreamService = make(map[string]*dto.UpstreamService)

// buffered 고루틴 (100개의 서비스는 일단 없는 것으로 간주)
// TODO 추후 Service 수가 많아질 경우, 정책 업데이트 시점에 일괄적으로 업데이트 하는 방식으로 변경 고려
type policyUpdateJob struct {
	service string
	payload string
}

type policyUpdateResult struct {
	service string
	policy  *dto.UpstreamService
}

var policyUpdateJobs = make(chan policyUpdateJob, 100)
var policyUpdateResults = make(chan policyUpdateResult, 100)

/*
초기화 시점에 GlideValkey를 통해 "UPSTREAM:*" 패턴의 키를 모두 조회하여 정책 맵을 초기화
- GlideValkey의 SCAN 명령을 사용하여 "UPSTREAM:*" 패턴의 키를 모두 조회
- 조회된 키에 대해 MGET 명령으로 일괄 조회하여 정책 데이터를 가져옴G
- 가져온 정책 데이터를 policyMap에 저장하여 서비스/호스트/URI별 정책을 관리
- 초기화 이후에도 GlideValkey의 Pub/Sub 기능을 활용하여 정책 변경 이벤트를 수신하고, 변경된 정책을 실시간으로 업데이트할 수 있도록 구현 예정
*/
func initPolocyChannel() {
	for i := 0; i < 50; i++ {
		go func(id int, job <-chan policyUpdateJob, result chan<- policyUpdateResult) {
			var upstreamService dto.UpstreamService
			item := <-job

			err := json.Unmarshal([]byte(item.payload), &upstreamService)
			if err != nil {
				log.Fatalf("Initialize policy error:policy unmarshal error: %v", err)
			}

			result <- policyUpdateResult{service: item.service, policy: &upstreamService}
		}(i, policyUpdateJobs, policyUpdateResults)
	}
}

func NewPolicyService(valkeyClient *infra.GlideValkey) *policyService {
	initPolocyChannel()
	initalizePolicyMap(valkeyClient.GetClient(), 0)

	return &policyService{valkeyClient: valkeyClient}
}

func initalizePolicyMap(valkey valkey.Client, cursor uint64) {
	keyCount := int64(1000)

	command := valkey.B().Scan().Cursor(cursor).Match("UPSTREAM:*").Count(keyCount).Build()
	result := valkey.Do(context.Background(), command)

	if result.Error() != nil {
		log.Fatalf("Initialize policy error: %v", result.Error())
	}

	scanEntity, err := result.AsScanEntry()
	if err != nil {
		log.Fatalf("Initialize policy error: %v", result.Error())
	}

	log.Printf("Scanned %d keys\n", len(scanEntity.Elements))

	nextCursor := scanEntity.Cursor
	keys := scanEntity.Elements

	if len(keys) == 0 {
		return
	}

	mgetCommdns := valkey.B().Mget().Key(keys...).Cache()
	mgetResult := valkey.DoCache(context.Background(), mgetCommdns, 365*time.Hour)

	if mgetResult.Error() != nil {
		log.Fatalf("Initialize policy error:mget error: %v", mgetResult.Error())
	}

	values, err := mgetResult.AsStrSlice()
	if err != nil {
		log.Fatalf("Initialize policy error:mget parse error: %v", mgetResult.Error())
	}

	// goroutine pool로 병렬 처리, 정책 업데이트 채널에 job 전달
	for index, value := range values {
		policyUpdateJobs <- policyUpdateJob{
			service: extractServiceName(keys[index]),
			payload: value,
		}
	}

	// 결과 수신
	for range values {
		result := <-policyUpdateResults
		result.policy.InitializeResourceIndex()
		policyMap[result.service] = result.policy
	}

	if nextCursor != 0 {
		initalizePolicyMap(valkey, nextCursor)
	}
}

func extractServiceName(rawKey string) string {
	return strings.TrimPrefix(rawKey, "UPSTREAM:")
}

func (p *policyService) CheckPolicy(urlParseDto modelDto.URLParseDTO) (modelDto.RewitePathDTO, error) {
	upstreamService, ok := policyMap[urlParseDto.Service]
	if !ok {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching service found for %s", urlParseDto.Service)
	}

	domain, emptyDomain := upstreamService.LookupResourceDomain(urlParseDto.Domain)
	if domain == nil {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching domain found for %s", urlParseDto.Domain)
	}

	lookupPath := urlParseDto.GetPath(emptyDomain)

	pathStream := domain.LookupPath(lookupPath)
	if pathStream == nil {
		return modelDto.NewEmptyRewitePathDTO(), fmt.Errorf("No matching path found for %s", lookupPath)
	}

	return modelDto.NewRewitePathDTO(pathStream), nil
}

var PolicyServiceSet = wire.NewSet(
	NewPolicyService,
	wire.Bind(new(PolicyService), new(*policyService)),
)
