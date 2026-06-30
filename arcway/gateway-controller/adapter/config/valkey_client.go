package config

import (
	"fmt"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type ValkeyClient struct {
	SingleClient valkey.Client
}

var VakeyClientSet = wire.NewSet(NewValkeyClient)

func NewValkeyClient(appConfig *AppConfig) *ValkeyClient {
	clientOption := valkey.ClientOption{
		InitAddress: appConfig.Valkey.Hosts,
		Standalone: valkey.StandaloneOption{
			EnableRedirect: true,
		},
	}

	client, err := valkey.NewClient(clientOption)
	if err != nil {
		panic(fmt.Errorf("valkey.NewClient: %w", err))
	}

	return &ValkeyClient{
		SingleClient: client,
	}
}
