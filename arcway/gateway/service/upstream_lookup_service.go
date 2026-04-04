package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	commonModel "gateway/common/model"
	"gateway/common/model/rewrite"
	"gateway/config"
	"gateway/gwe"
	"gateway/model"

	"github.com/google/wire"
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
			gwe.ErrLookupTransport,
			gwe.ErrMsgTransport,
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
			gwe.ErrLookupReadBody,
			gwe.ErrMsgReadBody,
			errorDetail,
		)
	}

	// httpStatus 에러가 있는 경우..
	if res.StatusCode != http.StatusOK {
		var errResponse *commonModel.ErrorResponse
		if err := json.Unmarshal(bodyBuffer, &errResponse); err != nil {
			errorDetail := fmt.Errorf("failed to unmarshal error response body: %v", err)

			log.Printf(
				"%v, targetPath: %s, responsebody: %s",
				errorDetail, targetPath, string(bodyBuffer),
			)

			return model.NewUpstreamLookupError(
				gwe.ErrLookupDecodeErrorBody,
				gwe.ErrMsgDecodeErrorBody,
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
			gwe.ErrLookupUpstreamResult,
			errResponse.Message,
			errors.New(errResponse.Detail),
		)
	}

	var pathInfo *rewrite.RewritePathDTO
	if err := json.Unmarshal(bodyBuffer, &pathInfo); err != nil {
		errorDetail := fmt.Errorf("failed to unmarshal response body: %v", err)
		log.Printf(
			"%v, targetPath: %s, response body: %s",
			errorDetail, targetPath, string(bodyBuffer),
		)

		return model.NewUpstreamLookupError(
			gwe.ErrLookupDecodeBody,
			gwe.ErrMsgDecodeBody,
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
		return nil, fmt.Errorf("%s: %v", gwe.ErrMsgTransport, err)
	}

	return bodyBuffer, nil
}

var UpstreamLookupServiceSet = wire.NewSet(
	NewUpstreamLookupService,
	wire.Bind(new(UpstreamLookupService), new(*upstreamLookupService)),
)
