package di

import (
	adapterGatewayController "gateway/adapter/out/gatewaycontroller"
	ddapterGatewayRateLimit "gateway/adapter/out/ratelimit"
	appPortOut "gateway/application/port/out"
	appPortOutRateLimiter "gateway/application/port/out/ratelimiter"

	"github.com/google/wire"
)

type AdapterOutDI struct {
	GatewayControllerClient appPortOut.UpstreamLookupPort
	RateLimiter             appPortOutRateLimiter.RateLimiterPort
}

var AdapterOutDiProviderSet = wire.NewSet(
	adapterGatewayController.NewUpstreamLookup,
	ddapterGatewayRateLimit.NewRateLimit,
	wire.Struct(new(AdapterOutDI), "*"),
)
