package router

import (
	"encoding/json"
	"fmt"
	"gateway/controller/model/dto"
	"gateway/controller/service"
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
	cr.Mux.HandleFunc("GET /v1/uri/policy", func(w http.ResponseWriter, r *http.Request) {
		targetUrl := r.URL.Query().Get("upstream")
		if targetUrl == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dto, err := parseTargetUrl(targetUrl)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		info, err := cr.service.GetRouteInfo(dto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)
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
