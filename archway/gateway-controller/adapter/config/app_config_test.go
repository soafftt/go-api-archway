package config

import "testing"

func TestNewAppConfigParsesValkeyMasterHost(t *testing.T) {
	t.Setenv("VALKEY_MASTER_HOST", "127.0.0.1:6379")

	cfg := NewAppConfig()

	if cfg.Valkey.MasterHost != "127.0.0.1:6379" {
		t.Fatalf("expected valkey master host %q, got %q", "127.0.0.1:6379", cfg.Valkey.MasterHost)
	}
}
