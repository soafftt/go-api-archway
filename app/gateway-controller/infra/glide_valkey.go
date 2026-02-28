package infra

import (
	"gateway/controller/config"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type GlideValkey struct {
	client *valkey.Client
}

// TODO: close client when app shutdown
func NewGlideValkey(config *config.AppConfig) *GlideValkey {
	valkeyConfig := config.Valkey

	client, err := valkey.NewClient(
		valkey.ClientOption{
			InitAddress: valkeyConfig.Hosts,
			Standalone: valkey.StandaloneOption{
				EnableRedirect: true,
			},
			ClientTrackingOptions: []string{"PREFIX", "UPSTREAM:"},
		},
	)

	// client connection error is fatal, panic here to fail fast
	if err != nil {
		panic(err)
	}

	return &GlideValkey{client: &client}
}

func (v *GlideValkey) GetClient() *valkey.Client {
	return v.client
}

var GlideValkeySet = wire.NewSet(NewGlideValkey)
