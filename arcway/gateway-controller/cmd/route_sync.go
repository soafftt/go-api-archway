package main

import (
	"context"
	cdp "core/domain/pubsub"
	"encoding/json"
	"log"

	"github.com/valkey-io/valkey-go"
)

type routeOperationMessage struct {
	Method  string `json:"method"`
	Service string `json:"service"`
}

func (a *GatewayControllerApp) LoadRouteCache() {
	a.adapterPortOut.RouterCache.RouteCache.LoadCache()
}

func (a *GatewayControllerApp) ListenRouteOperations(ctx context.Context, ready chan<- struct{}) error {
	var readyOnce bool
	ctx = valkey.WithOnSubscriptionHook(ctx, func(subscription valkey.PubSubSubscription) {
		if readyOnce {
			return
		}
		if subscription.Kind == "subscribe" && subscription.Channel == cdp.ROUTE_CHANNEL {
			readyOnce = true
			close(ready)
		}
	})
	command := a.valkeyClient.SingleClient.B().Subscribe().Channel(cdp.ROUTE_CHANNEL).Build()
	return a.valkeyClient.SingleClient.Receive(ctx, command, func(message valkey.PubSubMessage) {
		routeMessage := routeOperationMessage{}
		if err := json.Unmarshal([]byte(message.Message), &routeMessage); err != nil {
			log.Printf("route update unmarshal failed: %v", err)
			return
		}
		if routeMessage.Service == "" {
			return
		}

		switch routeMessage.Method {
		case cdp.ROUTE_MESSAGE_ADD, cdp.ROUTE_MESSAGE_UPDATE:
			if err := a.adapterPortOut.RouterCache.RouteCache.Update([]string{routeMessage.Service}); err != nil {
				log.Printf("route cache update failed for service %s: %v", routeMessage.Service, err)
			}
		case cdp.ROUTE_MESSAGE_DELETE:
			a.adapterPortOut.RouterCache.RouteCache.Evict(routeMessage.Service)
		default:
			log.Printf("unknown route operation method: %s", routeMessage.Method)
		}
	})
}
