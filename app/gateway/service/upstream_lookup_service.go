package service

import (
	"encoding/json"
	"errors"
	"fmt"
	commonModel "gateway/common/model"
	"gateway/common/model/rewrite"
	"gateway/config"
	"gateway/model"
	"io"
	"log"
	"net/http"

	"github.com/google/wire"
)

const (
	UNIX_SOCKET_RESPONSE_ERROR                               = "UNIX_SOCKET_RESPONSE_ERROR"
	UNIX_SOCKET_RESPONSE_BODY_ERR_READ_ERROR                 = "UNIX_SOCKET_RESPONSE_BODY_READ_ERROR"
	UNIX_SOCKET_RESPONSE_BODY_ERR_JSON_UNMARSHAL_ERROR       = "UNIX_SOCKET_RESPONSE_BODY_JSON_UNMARSHAL_ERROR"
	UNIX_SOCKET_RESPONSE_ERROR_BODY_ERR_JSON_UNMARSHAL_ERROR = "UNIX_SOCKET_RESPONSE_ERROR_BODY_ERR_JSON_UNMARSHAL_ERROR"
)

type UpstreamLookupService interface {
	Lookup(targetPath string) model.UpstreamLookupResult
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
func (s *upstreamLookupService) Lookup(targetPath string) model.UpstreamLookupResult {
	res, err := s.httpClient.Get(s.lookupUrl + targetPath)
	if err != nil {
		return model.NewUpstreamLookupError(
			model.LookupErrorTransport,
			UNIX_SOCKET_RESPONSE_ERROR,
			fmt.Errorf("failed to call upstream lookup service: %v", err),
		)
	}

	bodyBuffer, err := bodyRead(res)
	if err != nil {
		errorDetail := fmt.Errorf("failed to read response body: %v", err)

		log.Printf(
			"%v, targetPath: %s",
			errorDetail, targetPath,
		)

		return model.NewUpstreamLookupError(
			model.LookupErrorReadBody,
			UNIX_SOCKET_RESPONSE_BODY_ERR_READ_ERROR,
			errorDetail,
		)
	}

	// httpStatus 에러가 있는 경우..
	if res.StatusCode != http.StatusOK {
		var errResponse *commonModel.ErrorResponse
		if err := json.Unmarshal(bodyBuffer, &errResponse); err != nil {
			errorDetail := fmt.Errorf("failed to unmarshal error response body: eror%v", err)

			log.Printf(
				"%v, targetPath: %s, responsebody: %s",
				errorDetail, targetPath, string(bodyBuffer),
			)

			return model.NewUpstreamLookupError(
				model.LookupErrorDecodeErrorBody,
				UNIX_SOCKET_RESPONSE_ERROR_BODY_ERR_JSON_UNMARSHAL_ERROR,
				errorDetail,
			)
		}

		log.Printf(
			"unix-socket: %s, detail: %v, target: %s",
			errResponse.Message,
			errResponse.Detail,
			targetPath,
		)

		return model.NewUpstreamLookupError(
			model.LookupErrorUpstreamResult,
			errResponse.Message,
			errors.New(errResponse.Detail),
		)
	}

	var pathInfo *rewrite.RewritePathDTO
	if err := json.Unmarshal(bodyBuffer, &pathInfo); err != nil {
		errorDetail := fmt.Errorf("failed to unmarshal response body: eror%v", err)
		log.Printf(
			"%v, targetPath: %s, responsebody: %s",
			errorDetail, targetPath, string(bodyBuffer),
		)

		return model.NewUpstreamLookupError(
			model.LookupErrorDecodeBody,
			UNIX_SOCKET_RESPONSE_BODY_ERR_JSON_UNMARSHAL_ERROR,
			errorDetail,
		)
	}

	return model.NewUpstreamLookupResult(pathInfo)

}

// body 읽기 함수
func bodyRead(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	bodyBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(UNIX_SOCKET_RESPONSE_ERROR+": %v", err)
	}

	return bodyBuffer, nil
}

var UpstreamLookupServiceSet = wire.NewSet(
	NewUpstreamLookupService,
	wire.Bind(new(UpstreamLookupService), new(*upstreamLookupService)),
)
