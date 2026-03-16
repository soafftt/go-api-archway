package service

import (
	"encoding/json"
	"fmt"
	model "gateway/common/model/rewrite"
	"gateway/config"
	"net/http"

	"github.com/google/wire"
)

const (
	UNIX_SOCKET_RESPONSE_ERROR               = "UNIX_SOCKET_RESPONSE_ERROR :%v"
	UNIX_SOCKET_RESPONSE_BODY_ERR_READ_ERROR = "UNIX_SOCKET_RESPONSE_BODY_READ_ERROR :%v"
)

type UpstreamLookupService interface {
	Lookup(targetPath string) (*model.RewitePathDTO, error)
}

type upstreamLookupService struct {
	httpClient *http.Client
	lookupUrl  string
}

func NewUpstreamLookupService(config *config.AppConfig, httpClient *http.Client) *upstreamLookupService {
	return &upstreamLookupService{
		httpClient: httpClient,
		lookupUrl:  config.UpstreamLookup.BaseURL,
	}
}

/*
Upstream 에 대한 정보를 조회합니다.
*/
func (s *upstreamLookupService) Lookup(targetPath string) (*model.RewitePathDTO, error) {
	res, err := s.httpClient.Get(s.lookupUrl + targetPath)
	if err != nil {
		return nil, fmt.Errorf(UNIX_SOCKET_RESPONSE_ERROR, err)
	}

	bodyBuffer, err := bodyRead(res)
	if err != nil {
		return nil, err
	}

	var pathInfo *model.RewitePathDTO
	if err := json.Unmarshal(bodyBuffer, &pathInfo); err != nil {
		return nil, fmt.Errorf(UNIX_SOCKET_RESPONSE_BODY_ERR_READ_ERROR, err)
	}

	return pathInfo, nil
}

// body 읽기 함수
func bodyRead(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	bodyBuffer := make([]byte, 0)

	_, err := res.Body.Read(bodyBuffer)
	if err != nil {
		return nil, fmt.Errorf(UNIX_SOCKET_RESPONSE_BODY_ERR_READ_ERROR, err)
	}

	return bodyBuffer, nil
}

var UpstreamLookupServiceSet = wire.NewSet(
	NewUpstreamLookupService,
	wire.Bind(new(UpstreamLookupService), new(*upstreamLookupService)),
)
