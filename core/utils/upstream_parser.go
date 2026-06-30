package utils

import (
	"core/domain/upstream"
	"core/gjwt"
	"encoding/json"
	"fmt"
)

var chanSource chan ParseRequest = make(chan ParseRequest, 100)
var chanResult chan ParseResult = make(chan ParseResult, 100)

// 패키지 초기화를 통한 GoRoutine 초기화.
func init() {
	for i := 0; i < 50; i++ {
		go parse(i, chanSource, chanResult)
	}
}

/*
Valkey 에 등록된 정보를 객체로 전환하기 위한 요청 구조체.
*/
type ParseRequest struct {
	service string
	payload string
}

func NewParseRequest(service string, payload string) ParseRequest {
	return ParseRequest{
		service: service,
		payload: payload,
	}
}

/*
처리 결과 전환 결과
*/
type ParseResult struct {
	err     error
	service string
	result  *upstream.UpstreamService
}

/*
객체 변환 고루틴
*/
func parse(jobId int, source <-chan ParseRequest, result chan<- ParseResult) {
	for request := range source {
		service := &upstream.UpstreamService{}

		err := json.Unmarshal([]byte(request.payload), service)
		if err != nil {
			Error("failed to parse upstream payload", "job_id", jobId, "error", err)
			result <- ParseResult{
				err: fmt.Errorf("parse upstream payload: %w", err),
			}
			continue
		}

		result <- ParseResult{
			service: request.service,
			result:  service,
		}
	}
}

// JSON을 upstreamService 로 변환하고, 객체 초기화를 한다.
func ParseToUpstreamServiceWithInitialize(sources []ParseRequest) []*upstream.UpstreamService {
	parseResults := make([]*upstream.UpstreamService, 0, len(sources))

	for r := range sources {
		chanSource <- sources[r]
	}

	for range sources {
		parseResult := <-chanResult
		if parseResult.err != nil {
			// 에러 처리.
			Error("err json to upstream service", parseResult.err)
			panic(parseResult.err)
		}

		upstreamService := parseResult.result

		// UpStreamService 를 초기화 하여, 하위 Path 를 모두 정규화 한다
		upstreamService.InitializeResourceIndex()

		if upstreamService.Authorization != nil {
			jwtAuthorization := upstreamService.Authorization

			// authorization 처리를 위한 jwt 등록
			if err := gjwt.RegisterKeyByString(parseResult.service, jwtAuthorization.KeyData, gjwt.JSONKey, jwtAuthorization.Algorithm); err != nil {
				Error("err jwt key register", err)
				continue
			}
		}

		parseResults = append(parseResults, upstreamService)
	}

	return parseResults
}
