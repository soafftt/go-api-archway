package gatewaycontroller

import (
	coreAdapterIn "core/adapter/in"
	"core/errs"
	"core/utils"
	"errors"
	"gateway/adapter/config"
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	"net/http"
)

type upstreamLookup struct {
	uri    string
	client *http.Client
}

func NewUpstreamLookup(
	appConfig *config.AppConfig,
	gatewayControllerClient client.GatewayControllerClient,
) out.UpstreamLookupPort {
	return &upstreamLookup{
		uri:    appConfig.GatewayController.BaseURL,
		client: gatewayControllerClient.GetClient(),
	}
}

func (u upstreamLookup) GetUpstreamInfo(
	path string,
	accessToken *string,
) (out.UpStreamLookupPortResult, error) {
	var lookupResult out.UpStreamLookupPortResult

	req, err := http.NewRequest("GET", u.uri+path, nil)
	if err != nil || req == nil {
		// 에러 처리.
		return lookupResult, errors.Join(errs.ERR_GATEWAY_CONTROLLER_SEND, err)
	}

	if accessToken != nil {
		req.Header.Add("Authorization", "Bearer "+*accessToken)
	}

	httpResponse, err := u.client.Do(req)
	if err != nil {
		return lookupResult, errors.Join(errs.ERR_GATEWAY_CONTROLLER_SEND, err)
	}

	defer func() {
		_ = httpResponse.Body.Close()
	}()

	// TODO:  에러 핸들링 하는 유틸을 만들어야 함.
	if httpResponse.StatusCode != http.StatusOK {
		errorResponse, err := utils.ToStruct[coreAdapterIn.ErrorResponse](httpResponse)
		if err != nil {
			return lookupResult, errors.Join(errs.ERR_GATEWAY_CONTROLLER_SEND, err)
		}

		// logging
		return lookupResult, errs.ToArchwayError(errorResponse.Message)
	}

	lookupResult, err = utils.ToStruct[out.UpStreamLookupPortResult](httpResponse)
	if err != nil {
		return lookupResult, errors.Join(errs.ERR_GATEWAY_CONTROLLER_SEND, err)
	}

	return lookupResult, nil
}
