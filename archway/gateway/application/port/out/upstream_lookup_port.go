package out

type UpstreamLookupPort interface {
	GetUpstreamInfo(path string, accessToken *string) (UpStreamLookupPortResult, error)
}

type UpstreamLookupGrpcPort UpstreamLookupPort
