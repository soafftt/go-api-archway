package infra

import "github.com/valkey-io/valkey-go"

type GlideValkey struct {
	Client *valkey.Client
}

func NewGlideValkey() *GlideValkey {
	valkey.NewClient(
		valkey.ClientOption{
			InitAddress: []string{"127.0.0.1:6379"},
		},
	)

	return &GlideValkey{}
}
