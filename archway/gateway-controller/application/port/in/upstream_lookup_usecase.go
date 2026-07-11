package in

import "gateway/controller/application/port/in/dto"

type UpstreamLookupUseCase interface {
	LookUp(dto dto.UpStreamLookupRequest) dto.UpStreamLookupResult
}
