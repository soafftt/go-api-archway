package outbound

import "gateway/model"

type UpstreamLookupPort interface {
	Lookup(targetPath string) model.UpstreamLookupResult
}
