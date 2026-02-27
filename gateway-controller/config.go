package main

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/google/wire"
	"github.com/joho/godotenv"
)

type Config struct {
	Valkey struct {
		Hosts    []string `env:"VALKEY_HOSTS,required" envSeparator:","`
		ReadFrom string   `env:"VALKEY_READFROM" envDefault:"primary"`
	}
}

func NewConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	for _, host := range cfg.Valkey.Hosts {
		if strings.TrimSpace(host) == "" {
			panic(fmt.Errorf("VALKEY_HOSTS contains an empty host"))
		}
	}

	return cfg
}

var ConfigSet = wire.NewSet(NewConfig)
