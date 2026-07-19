package app_config_test

import (
	"gateway/controller/adapter/config/app_config"
	"testing"
)

func TestNewAppConfigParsesValkeyMasterHost(t *testing.T) {
	t.Setenv("VALKEY_MASTER_HOST", "127.0.0.1:6379")

	cfg := app_config.NewAppConfig()

	if cfg.Valkey.MasterHost != "127.0.0.1:6379" {
		t.Fatalf("expected valkey master host %q, got %q", "127.0.0.1:6379", cfg.Valkey.MasterHost)
	}
}
