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
		// path 가 uri 형식이 아님.
		return NewEmptyUpStreamLookupDto(), err
	}

	// uri 를 "/" 로 구분함.
	segments := strings.Split(strings.Trim(uri.Path, "/"), "/")

	// 하단은 URI 규칙임.
	// e.g. v1/member/user/{1} 과 같거나.
	// e.g. v1/member/users/abcdef/dddd
	version := segments[0]
	service := segments[1]
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

// dmain 을 찾으면, path는 도메인 이후의 경로가 되어야 하기 때문에, 이와 같이 구현
func (u UpStreamLookupRequest) GetRelativePath(isEmptyDomain bool) string {
	if isEmptyDomain {
		return u.Path
	}

	return strings.Join(strings.Split(u.Path, "/")[1:], "/")
}

func (u UpStreamLookupRequest) String() string {
	return fmt.Sprintf("version=%s, service=%s, domain=%s, path=%s", u.Version, u.Service, u.Domain, u.Path)
}
