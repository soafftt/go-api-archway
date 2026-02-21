package main

type Config struct {
	Valkey struct {
		Hosts []string `env:"VALKEY_HOSTS" envSeparator:","`
	}
}
