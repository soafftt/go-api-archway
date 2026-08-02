package di

import (
	"gateway/adapter/out/controlplane/grpc"
	"gateway/adapter/out/controlplane/http"
	ddapterGatewayRateLimit "gateway/adapter/out/ratelimit"
	appPortOut "gateway/application/port/out"
	appPortOutRateLimiter "gateway/application/port/out/ratelimiter"

	"github.com/google/wire"
)

type AdapterOutDI struct {
	HttpUpstreamLookupPort appPortOut.UpstreamLookupPort
	GrpcUpstreamLookupPort appPortOut.UpstreamLookupGrpcPort
	RateLimiter            appPortOutRateLimiter.RateLimiterPort
	GrpcMetricLookupPort   grpc.GrpcMetricOutPort
	UnixMetricLookupPort   http.HttpMetricOutPort
}

var AdapterOutDiProviderSet = wire.NewSet(
	http.NewUpstreamLookup,
	grpc.NewUpstreamLookup,
	grpc.NewMetricsLookup,
	http.NewHttpMetricLookup,
	ddapterGatewayRateLimit.NewRateLimit,
	wire.Struct(new(AdapterOutDI), "*"),
)
