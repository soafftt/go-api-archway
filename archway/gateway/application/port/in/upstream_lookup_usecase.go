package in

type UpstreamLookupUseCase interface {
	Lookup(srcPath string, accessToken *string) (UpstreamLookupResult, error)
}
