package config

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
)

func TestTLSDefaultsUseKloudliteDirectory(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	unsetEnv(t, "TLS_CERT_FILE")
	unsetEnv(t, "TLS_KEY_FILE")

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.TLS.CertFile != "/etc/kloudlite/tls.crt" {
		t.Fatalf("TLS cert file = %q, want /etc/kloudlite/tls.crt", cfg.TLS.CertFile)
	}
	if cfg.TLS.KeyFile != "/etc/kloudlite/tls.key" {
		t.Fatalf("TLS key file = %q, want /etc/kloudlite/tls.key", cfg.TLS.KeyFile)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	previous, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	})
}
