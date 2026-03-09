package component

import (
	"encoding/json"
	dto "gateway/common/dto/upstream"
	"log"
)

// RouteSource는 파싱할 서비스명과 페이로드를 담는 구조체
type RouteSource struct {
	service string
	payload string
}

func NewRouteSource(service, payload string) RouteSource {
	return RouteSource{service: service, payload: payload}
}

// routeParseResult는 파싱 결과로 서비스명과 UpstreamService 구조체를 담는 구조체
type routeParseResult struct {
	service string
	policy  *dto.UpstreamService
}

type RouteParse interface {
	ParseFromSlice(sources []RouteSource) ([]*dto.UpstreamService, error)
	Parse(source RouteSource) (*dto.UpstreamService, error)
}

type routeParser struct {
	routeParseJobs    chan RouteSource
	routeParseResults chan routeParseResult
}

func NewRouteParser() *routeParser {
	parser := &routeParser{
		routeParseJobs:    make(chan RouteSource, 100),
		routeParseResults: make(chan routeParseResult, 100),
	}
	parser.initRouteParserChannel()

	return parser
}

// 파싱을 위한 코류틴 채널 초기화
func (r *routeParser) initRouteParserChannel() {
	for i := 0; i < 50; i++ {
		go parse(i, r.routeParseJobs, r.routeParseResults)
	}
}

// 파싱 코루틴 함수
func parse(id int, job <-chan RouteSource, result chan<- routeParseResult) {
	var upstreamService dto.UpstreamService
	item := <-job

	err := json.Unmarshal([]byte(item.payload), &upstreamService)
	if err != nil {
		log.Fatalf("Initialize policy error:policy unmarshal error: %v", err)

	}

	result <- routeParseResult{service: item.service, policy: &upstreamService}
}

// map[string]string 형태의 서비스명과 페이로드를 입력받아, 각 페이로드를 UpstreamService 구조체로 파싱하여 반환
func (r *routeParser) ParseFromSlice(sources []RouteSource) ([]*dto.UpstreamService, error) {
	for _, item := range sources {
		r.routeParseJobs <- item
	}

	upstreamServices := make([]*dto.UpstreamService, 0)
	for range sources {
		result := <-r.routeParseResults
		result.policy.InitializeResourceIndex()

		upstreamServices = append(upstreamServices, result.policy)
	}

	return upstreamServices, nil
}

func (r *routeParser) Parse(source RouteSource) (*dto.UpstreamService, error) {
	r.routeParseJobs <- source

	result := <-r.routeParseResults
	result.policy.InitializeResourceIndex()

	return result.policy, nil
}
