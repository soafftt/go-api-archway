package component

import (
	"context"
	"encoding/json"
	dto "gateway/common/model/pubsub"
	"gateway/controller/infra"
	"log"

	"github.com/valkey-io/valkey-go"
)

/*
Route 정보가 변경되거나 추가될때 Valkey 의 "UPSTREAM:*" 패턴의 키를 조회하여 정책 데이터를 가져와 RouteCache에 저장하는 역할
*/
type RouteMessageHook struct{}

func NewRouteMessageHook(rc RouteCache, valkeyWrap *infra.ValkeyWrap) RouteMessageHook {
	go hookPubSub(rc, valkeyWrap)
	return RouteMessageHook{}
}

func hookPubSub(rc RouteCache, valkeyWrap *infra.ValkeyWrap) {
	for {
		valkeyWrap.PubSubClient.Do(
			context.Background(),
			valkeyWrap.PubSubClient.B().Subscribe().Channel(dto.ROUTE_CHANNEL).Build(),
		)

		chanErr := valkeyWrap.PubSubClient.SetPubSubHooks(
			valkey.PubSubHooks{
				OnMessage: func(m valkey.PubSubMessage) {
					// 메세지 처리.
					go func() {
						log.Println(m.Message)

						var msg dto.RoutePubSubMessage
						if err := json.Unmarshal([]byte(m.Message), &msg); err != nil {
							log.Printf("Failed to unmarshal message: %v", err)
							return
						}

						// 수정 및 추가.
						if msg.Method == dto.ROUTE_MESSAGE_ADD || msg.Method == dto.ROUTE_MESSAGE_UPDATE {
							rc.Update(context.Background(), []string{msg.Service})
						} else {
							rc.Evict(msg.Service)
						}
					}()

				},
				OnSubscription: func(s valkey.PubSubSubscription) {
					log.Printf("Subscribed to channel: %s", s.Channel)
				},
			},
		)

		err := <-chanErr
		if err != nil {
			log.Printf("PubSub Channel Error occurred (Restart()): %v", err)
		}
	}

}
