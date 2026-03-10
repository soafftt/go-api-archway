package service

import (
	"encoding/json"
	dto "gateway/common/dto/upstream"
	"net/http"
)

type RouterCheckService struct {
	httpClient *http.Client
}

func NewRouterCheckService(httpClient *http.Client) *RouterCheckService {
	return &RouterCheckService{
		httpClient: httpClient,
	}
}

/*
Upstream 에 대한 정보를 조회합니다.
*/
func (s *RouterCheckService) CheckRoute(targetPath string) (*dto.UpstreamPath, error) {
	res, err := s.httpClient.Get("http://localhost/v1/upstream?path=" + targetPath)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bodyBuffer := make([]byte, 0)

	_, err = res.Body.Read(bodyBuffer)
	if err != nil {
		return nil, err
	}

	var pathInfo *dto.UpstreamPath

	err = json.Unmarshal(bodyBuffer, &pathInfo)
	if err != nil {
		return nil, err
	}

	return pathInfo, nil
}
