package in

import "gateway/controller/application/port/in/dto"

type UpstreamLookupUseCase interface {
	LookUpFromRequest(dto dto.UpStreamLookupRequest) dto.UpStreamLookupResult
	Lookup(
		Version string,
		Service string,
		Domain string,
		Path string,
		AccessToken *string,
	) dto.UpStreamLookupResult
}
