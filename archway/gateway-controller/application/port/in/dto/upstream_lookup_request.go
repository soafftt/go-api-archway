package dto

import (
	"core/errs"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type UpStreamLookupRequest struct {
	Version     string
	Service     string
	Domain      string
	Path        string
	AccessToken *string
}

func NewEmptyUpStreamLookupDto() UpStreamLookupRequest {
	var request UpStreamLookupRequest
	return request
}

func NewUpStreamLookupRequest(r *http.Request) (UpStreamLookupRequest, error) {
	targetUrl := r.URL.Query().Get("path")
	if targetUrl == "" {
		// 에러 응답 : targetUrl 이 없음.
		return NewEmptyUpStreamLookupDto(), errs.ERR_INVALID_TARGET
	}

	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	authorization = strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))

	var accessToken *string
	if authorization != "" {
		accessToken = &authorization
	}

	request, err := parseRewritePath(targetUrl, accessToken)
	if err != nil {
		return NewEmptyUpStreamLookupDto(), err
	}

	return request, nil
}

func parseRewritePath(path string, accessToken *string) (UpStreamLookupRequest, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return NewEmptyUpStreamLookupDto(), err
	}

	segments := strings.Split(strings.Trim(uri.Path, "/"), "/")
	if len(segments) < 4 {
		return NewEmptyUpStreamLookupDto(), errs.ERR_INVALID_TARGET
	}

	serviceIndex := 0
	versionIndex := 1
	if isVersionSegment(segments[0]) {
		serviceIndex = 1
		versionIndex = 0
	}

	service := segments[serviceIndex]
	version := segments[versionIndex]
	resourceDomain := segments[2]
	resourcePath := strings.Join(segments[2:], "/")

	return UpStreamLookupRequest{
		Version:     version,
		Service:     service,
		Domain:      resourceDomain,
		Path:        resourcePath,
		AccessToken: accessToken,
	}, nil
}

func isVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	for _, character := range segment[1:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func (u UpStreamLookupRequest) GetRelativePath(isEmptyDomain bool) string {
	if isEmptyDomain {
		return u.Path
	}

	return strings.Join(strings.Split(u.Path, "/")[1:], "/")
}

func (u UpStreamLookupRequest) String() string {
	return fmt.Sprintf("version=%s, service=%s, domain=%s, path=%s", u.Version, u.Service, u.Domain, u.Path)
}
