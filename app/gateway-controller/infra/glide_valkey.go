package infra

import (
	"gateway/controller/config"
	"log"

	"github.com/google/wire"
	"github.com/valkey-io/valkey-go"
)

type GlideValkey struct {
	client valkey.Client // valkey client instance
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
			DisableCache:          false,
			ClientTrackingOptions: []string{"BCAST", "PREFIX", "UPSTREAM:"},
			AlwaysRESP2:           false,
		},
	)

	// client connection error is fatal, panic here to fail fast
	if err != nil {
		log.Fatalf("valkey glide client init fail: %v", err)
	}

	return &GlideValkey{client: client}
}

func (v *GlideValkey) GetClient() valkey.Client {
	return v.client
}

var GlideValkeySet = wire.NewSet(NewGlideValkey)
