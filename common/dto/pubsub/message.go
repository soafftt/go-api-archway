package pubsub

const (
	ROUTE_CHANNEL        = "ROUTE_OPERATIONS"
	ROUTE_MESSAGE_ADD    = "ROUTE_MESSAGE_ADD"
	ROUTE_MESSAGE_UPDATE = "ROUTE_MESSAGE_UPDATE"
	ROUTE_MESSAGE_DELETE = "ROUTE_MESSAGE_DELETE"
)

type RoutePubSubMessage struct {
	Method  string `json:"method"`
	Service string `json:"service"`
}
