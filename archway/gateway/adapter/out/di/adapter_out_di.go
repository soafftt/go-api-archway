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
	GrpcMetricLookupPort   controlPlane.GrpcMetricOutPort
	UnixMetricLookupPort   controlPlane.UnixMetricOutPort
}

var AdapterOutDiProviderSet = wire.NewSet(
	controlPlane.NewUpstreamLookup,
	controlPlane.NewUpstreamLookupGrpc,
	controlPlane.NewMetricsLookup,
	controlPlane.NewUnixMetricLookup,
	ddapterGatewayRateLimit.NewRateLimit,
	wire.Struct(new(AdapterOutDI), "*"),
)
