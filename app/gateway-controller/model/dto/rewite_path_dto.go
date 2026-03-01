package dto

import upstreamDto "gateway/common/dto/upstream"

type RewitePathDTO struct {
	Host               string `json:"domain"`              // 도메인 이름
	Path               string `json:"path"`                // 프록시 경로
	Method             string `json:"method"`              // 메소드
	ResponseTimeout    int64  `json:"response_timeout"`    // 응답 타임아웃
	RequestTimeout     int64  `json:"request_timeout"`     // 요청 타임아웃
	CheckAuthorization bool   `json:"check_authorization"` // 권한 체크 여부
	CacheTimeout       int64  `json:"cache_timeout"`       // 캐시 타임아웃
}

func NewEmptyRewitePathDTO() RewitePathDTO {
	return RewitePathDTO{}
}

func NewRewitePathDTO(upstreamPath *upstreamDto.UpstreamPath) RewitePathDTO {
	return RewitePathDTO{
		Host:               upstreamPath.Path,
		Path:               upstreamPath.Path,
		Method:             upstreamPath.Method,
		ResponseTimeout:    upstreamPath.ResponseTimeout,
		RequestTimeout:     upstreamPath.RequestTimeout,
		CheckAuthorization: upstreamPath.CheckAuthorization,
		CacheTimeout:       upstreamPath.CacheTimeout,
	}
}
