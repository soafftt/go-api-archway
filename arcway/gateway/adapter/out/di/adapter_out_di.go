package di

import (
	adapterGatewayController "gateway/adapter/out/gatewaycontroller"
	appPortOut "gateway/application/port/out"

	"github.com/google/wire"
)

type AdapterOutDI struct {
	GatewayControllerClient appPortOut.UpstreamLookupPort
}

var AdapterOutDiProviderSet = wire.NewSet(
	adapterGatewayController.NewUpstreamLookup,
	wire.Struct(new(AdapterOutDI), "*"),
)
