package valkey

import (
	"fmt"
	"gateway/controller/adapter/config/app_config"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type ValkeyClient interface {
	GetClient() valkey.Client
	GetSubscribeClient() valkey.Client
}

type valkeyClient struct {
	singleClient    valkey.Client
	subscribeClient valkey.Client
}

var ValkeyClientSet = wire.NewSet(NewValkeyClient)

func NewValkeyClient(appConfig *app_config.AppConfig) ValkeyClient {
	masterHost := appConfig.Valkey.MasterHost
	replicaHosts := appConfig.Valkey.ReplicaHosts

	singleClient := makeValkeyClient(
		masterHost,
		replicaHosts,
		func(clientOption valkey.ClientOption) {
			clientOption.ClientTrackingOptions = []string{"BCAST", "PREFIX", "UPSTREAM:"}
		},
	)

	subscribeClient := makeValkeyClient(
		masterHost,
		replicaHosts,
		nil,
	)

	return &valkeyClient{
		singleClient:    singleClient,
		subscribeClient: subscribeClient,
	}
}

func makeValkeyClient(
	masterHost string,
	replicaHosts []string,
	addOptionFn func(clientOption valkey.ClientOption),
) valkey.Client {
	standAloneOption := valkey.StandaloneOption{
		EnableRedirect: true,
	}

	if replicaHosts != nil && len(replicaHosts) > 0 {
		standAloneOption.ReplicaAddress = replicaHosts
	}

	clientOption := valkey.ClientOption{
		InitAddress: []string{masterHost},
		Standalone:  standAloneOption,
	}

	if addOptionFn != nil {
		addOptionFn(clientOption)
	}

	client, err := valkey.NewClient(clientOption)
	if err != nil {
		panic(fmt.Errorf("valkey.NewClient: %w", err))
	}

	return client
}

func (c *valkeyClient) GetClient() valkey.Client {
	return c.singleClient
}

func (c *valkeyClient) GetSubscribeClient() valkey.Client {
	return c.subscribeClient
}
