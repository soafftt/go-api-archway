package in

type Transport string

const (
	UnixHttp Transport = "http"
	UnixGrpc Transport = "grpc"
)

type UpstreamLookupUseCase interface {
	Lookup(srcPath string, accessToken *string, transport Transport) (UpstreamLookupResult, error)
}
