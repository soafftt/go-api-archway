package router

import (
	"fmt"
	"gateway/controller/service"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/wire"
)

type ControllerRouter struct {
	Mux     *http.ServeMux
	service service.PolicyService
}

func NewControllerRouter(policyService service.PolicyService) *ControllerRouter {
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

		service, path, err := parseTargetUrl(targetUrl)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cr.service.CheckPolicy(service, path)
	})
}

func parseTargetUrl(rawUrl string) (service string, path string, err error) {
	uri, err := url.Parse(rawUrl)
	if err != nil {
		return "", "", fmt.Errorf("Target UriParse Error: %v", err)
	}

	uriPath := strings.Split(uri.Path, "/")

	service = uriPath[0]                  // assuming the first segment of the path is the service name
	path = strings.Join(uriPath[1:], "/") // the rest is the service path

	return service, path, nil
}

var RouterSet = wire.NewSet(NewControllerRouter)
