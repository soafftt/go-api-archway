package dto

import "strings"

type URLParseDTO struct {
	Version string
	Service string
	Domain  string
	path    string
}

func NewEmptyURLParseDTO() URLParseDTO {
	return URLParseDTO{}
}

func NewURLParseDTO(version, service, domain, path string) URLParseDTO {
	return URLParseDTO{
		Version: version,
		Service: service,
		Domain:  domain,
		path:    path,
	}
}

// dmain 을 찾으면, path는 도메인 이후의 경로가 되어야 하기 때문에, 이와 같이 구현
func (u URLParseDTO) GetPath(isEmptyDomain bool) string {
	if isEmptyDomain {
		return u.path
	}

	return strings.Join(strings.Split(u.path, "/")[1:], "/")
}
