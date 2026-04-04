package infra

import (
	"log"

	"gateway/controller/config"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type ValkeyWrap struct {
	client       valkey.Client          // valkey client instance
	PubSubClient valkey.DedicatedClient // pub/sub 전용 valkey client instance
}

// 현재는 valkey-go 라이브러리를 사용하고 있지만 glide-valkey 로 변경해야 함.
func NewValkeyWrap(config *config.AppConfig) *ValkeyWrap {
	client := newValkeyClient(
		config,
		func(clientOption valkey.ClientOption) valkey.ClientOption {
			clientOption.DisableCache = false
			clientOption.ClientTrackingOptions = []string{"BCAST", "PREFIX", "UPSTREAM:"}
			clientOption.AlwaysRESP2 = false

			return clientOption
		},
	)

	pubsubDedicateClient, _ := newValkeyClient(
		config,
		func(clientOption valkey.ClientOption) valkey.ClientOption {
			return clientOption
		},
	).Dedicate()

	return &ValkeyWrap{client: client, PubSubClient: pubsubDedicateClient}
}

func (v *ValkeyWrap) GetClient() valkey.Client {
	return v.client
}

func newValkeyClient(
	config *config.AppConfig,
	fnSetClientOption func(valkey.ClientOption) valkey.ClientOption,
) valkey.Client {
	valkeyConfig := config.Valkey

	clientOption := valkey.ClientOption{
		InitAddress: valkeyConfig.Hosts,
		Standalone: valkey.StandaloneOption{
			EnableRedirect: true,
		},
	}
	clientOption = fnSetClientOption(clientOption)
	client, err := valkey.NewClient(clientOption)
	// client connection error is fatal, panic here to fail fast
	if err != nil {
		log.Fatalf("valkey glide client init fail: %v", err)
	}

	return client
}

var GlideValkeySet = wire.NewSet(NewValkeyWrap)
