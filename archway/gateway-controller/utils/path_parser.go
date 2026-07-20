package utils

import (
	"core/errs"
	"gateway/controller/application/port/in/dto"
	"net/url"
	"strings"
)

func ParseRewritePath(path string, accessToken *string) (dto.UpStreamLookupRequest, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return dto.NewEmptyUpStreamLookupDto(), err
	}

	segments := strings.Split(strings.Trim(uri.Path, "/"), "/")
	if len(segments) < 4 {
		return dto.NewEmptyUpStreamLookupDto(), errs.ERR_INVALID_TARGET
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

	return dto.UpStreamLookupRequest{
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
