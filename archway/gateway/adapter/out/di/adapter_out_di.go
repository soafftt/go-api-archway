package di

import (
	controlPlane "gateway/adapter/out/controlplane"
	ddapterGatewayRateLimit "gateway/adapter/out/ratelimit"
	appPortOut "gateway/application/port/out"
	appPortOutRateLimiter "gateway/application/port/out/ratelimiter"

	"github.com/google/wire"
)

type AdapterOutDI struct {
	HttpUpstreamLookupPort appPortOut.UpstreamLookupPort
	GrpcUpstreamLookupPort appPortOut.UpstreamLookupGrpcPort
	RateLimiter            appPortOutRateLimiter.RateLimiterPort
}

var AdapterOutDiProviderSet = wire.NewSet(
	controlPlane.NewUpstreamLookup,
	controlPlane.NewUpstreamLookupGrpc,
	ddapterGatewayRateLimit.NewRateLimit,
	wire.Struct(new(AdapterOutDI), "*"),
)
