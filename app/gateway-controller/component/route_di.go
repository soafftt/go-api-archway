package component

import "github.com/google/wire"

type ComponentSet struct {
	RouteParser      RouteParse
	RouteCache       RouteCache
	RouteMessageHook RouteMessageHook
}

var RouteComponentSet = wire.NewSet(
	NewRouteParser,
	NewUpstreamRouteCache,
	wire.Bind(new(RouteParse), new(*routeParser)),
	wire.Bind(new(RouteCache), new(*routeCache)),
	NewRouteMessageHook,
	wire.Struct(new(ComponentSet), "*"),
)
