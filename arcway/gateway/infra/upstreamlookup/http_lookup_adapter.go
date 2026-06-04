package upstreamlookup

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
	"gateway/port/outbound"

	"github.com/google/wire"
)

type HTTPUpstreamLookupAdapter struct {
	httpClient *http.Client
	lookupURL  string
}

func NewHTTPUpstreamLookupAdapter(config *config.AppConfig, httpClient *http.Client) *HTTPUpstreamLookupAdapter {
	return &HTTPUpstreamLookupAdapter{
		httpClient: httpClient,
		lookupURL:  config.UpstreamLookup.BaseURL,
	}
}

func (a *HTTPUpstreamLookupAdapter) Lookup(targetPath string) model.UpstreamLookupResult {
	res, err := a.httpClient.Get(a.lookupURL + targetPath)
	if err != nil {
		return model.NewUpstreamLookupError(
			gwe.ErrLookupTransport,
			gwe.ErrMsgTransport,
			fmt.Errorf("failed to call upstream lookup service: %v", err),
		)
	}

	bodyBuffer, err := readBody(res)
	if err != nil {
		errorDetail := fmt.Errorf("failed to read response body: %v", err)
		log.Printf("%v, targetPath: %s", errorDetail, targetPath)

		return model.NewUpstreamLookupError(
			gwe.ErrLookupReadBody,
			gwe.ErrMsgReadBody,
			errorDetail,
		)
	}

	if res.StatusCode != http.StatusOK {
		var errResponse *commonModel.ErrorResponse
		if err := json.Unmarshal(bodyBuffer, &errResponse); err != nil {
			errorDetail := fmt.Errorf("failed to unmarshal error response body: %v", err)
			log.Printf("%v, targetPath: %s, responsebody: %s", errorDetail, targetPath, string(bodyBuffer))

			return model.NewUpstreamLookupError(
				gwe.ErrLookupDecodeErrorBody,
				gwe.ErrMsgDecodeErrorBody,
				errorDetail,
			)
		}

		log.Printf("unix-socket: %s, detail: %v, target: %s", errResponse.Message, errResponse.Detail, targetPath)

		return model.NewUpstreamLookupError(
			gwe.ErrLookupUpstreamResult,
			errResponse.Message,
			errors.New(errResponse.Detail),
		)
	}

	var pathInfo *rewrite.RewritePathDTO
	if err := json.Unmarshal(bodyBuffer, &pathInfo); err != nil {
		errorDetail := fmt.Errorf("failed to unmarshal response body: %v", err)
		log.Printf("%v, targetPath: %s, response body: %s", errorDetail, targetPath, string(bodyBuffer))

		return model.NewUpstreamLookupError(
			gwe.ErrLookupDecodeBody,
			gwe.ErrMsgDecodeBody,
			errorDetail,
		)
	}

	return model.NewUpstreamLookupResult(pathInfo)
}

func readBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	bodyBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", gwe.ErrMsgTransport, err)
	}

	return bodyBuffer, nil
}

var UpstreamLookupAdapterSet = wire.NewSet(
	NewHTTPUpstreamLookupAdapter,
	wire.Bind(new(outbound.UpstreamLookupPort), new(*HTTPUpstreamLookupAdapter)),
)
