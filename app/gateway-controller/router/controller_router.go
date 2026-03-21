package router

import (
	"encoding/json"
	"fmt"
	commonCode "gateway/common/code"
	"gateway/common/model"
	"gateway/controller/model/dto"
	"gateway/controller/service"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/wire"
)

type ControllerRouter struct {
	Mux     *http.ServeMux
	service service.RouteService
}

func NewControllerRouter(policyService service.RouteService) *ControllerRouter {
	router := &ControllerRouter{
		Mux:     http.NewServeMux(),
		service: policyService,
	}

	router.registerRoutes()

	return router
}

func (cr *ControllerRouter) registerRoutes() {
	cr.Mux.HandleFunc("GET /v1/upstream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		targetUrl := r.URL.Query().Get("path")
		if targetUrl == "" {
			log.Println("not include targetUrl.")

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErroeResponse{Message: commonCode.ERROR_NOT_FOUND_PARAMETER, Detail: "missing required parameter: path"})

			return
		}

		dto, err := parseTargetUrl(targetUrl)
		if err != nil {
			log.Printf("targetUrl parsing error %s, %v", targetUrl, err)

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErroeResponse{Message: commonCode.ERROR_TARGET_URL_PARSING_FAILED, Detail: err.Error()})

			return
		}

		result := cr.service.GetRouteInfo(dto)
		if !result.Ok {
			detailError := result.Error.Detail

			log.Printf("upstrem check error : %v", detailError.Error())

			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(model.ErroeResponse{Message: result.Error.Message, Detail: detailError.Error()})

			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result.RewitePath)
	})
}

func parseTargetUrl(rawUrl string) (urlParseDto dto.URLParseDTO, err error) {
	uri, err := url.Parse(rawUrl)
	if err != nil {
		return dto.NewEmptyURLParseDTO(), fmt.Errorf("Target UriParse Error: %v", err)
	}

	segments := strings.Split(strings.Trim(uri.Path, "/"), "/")
	if len(segments) < 3 {
		return dto.NewEmptyURLParseDTO(), fmt.Errorf("invalid upstream path: %s", uri.Path)
	}

	version := segments[0]
	service := segments[1]
	resourceDomain := segments[2]
	resourcePath := strings.Join(segments[2:], "/")

	urlParseDto = dto.NewURLParseDTO(
		version,
		service,
		resourceDomain,
		resourcePath,
	)

	return urlParseDto, nil
}

var RouterSet = wire.NewSet(NewControllerRouter)
