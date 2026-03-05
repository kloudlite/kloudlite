package wmingress

import (
	"crypto/tls"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestTLSManagerConcurrentUpdates(t *testing.T) {
	logger := zap.NewNop()

	manager := NewTLSManager(logger)

	// Test 1: Concurrent certificate updates
	t.Run("ConcurrentCertificateUpdates", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50
		updatesPerGoroutine := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < updatesPerGoroutine; j++ {
					certificates := map[string]*TLSCertificate{
						"example.com": {
							Hosts:    []string{"example.com"},
							CertPEM:  []byte("cert-data"),
							KeyPEM:   []byte("key-data"),
							SecretID: "secret-1",
						},
					}

					if err := manager.UpdateCertificates(certificates); err != nil {
						t.Errorf("UpdateCertificates failed: %v", err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify final state is consistent
		metrics := manager.GetMetrics()
		if count, ok := metrics["certificate_count"].(int); ok {
			if count != 1 {
				t.Errorf("Expected 1 certificate, got %d", count)
			}
		} else {
			t.Error("certificate_count not found in metrics")
		}
	})

	// Test 2: Concurrent updates and reads
	t.Run("ConcurrentUpdatesAndReads", func(t *testing.T) {
		var wg sync.WaitGroup
		numUpdateGoroutines := 10
		numReadGoroutines := 50

		// Start update goroutines
		for i := 0; i < numUpdateGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 100; j++ {
					certificates := map[string]*TLSCertificate{
						"test.com": {
							Hosts:    []string{"test.com"},
							CertPEM:  []byte("cert-data"),
							KeyPEM:   []byte("key-data"),
							SecretID: "secret-1",
						},
					}

					if err := manager.UpdateCertificates(certificates); err != nil {
						t.Errorf("UpdateCertificates failed: %v", err)
					}
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Start read goroutines
		for i := 0; i < numReadGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < 100; j++ {
					_ = manager.GetCertificates()
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		wg.Wait()

		// Verify final state is consistent
		certs := manager.GetCertificates()
		if len(certs) != 1 {
			t.Errorf("Expected 1 certificate, got %d", len(certs))
		}
	})
}

func TestTLSManagerGetCertificatesImmutability(t *testing.T) {
	logger := zap.NewNop()

	manager := NewTLSManager(logger)

	// Set initial certificates
	initialCerts := map[string]*TLSCertificate{
		"example.com": {
			Hosts:    []string{"example.com"},
			CertPEM:  []byte("cert-data"),
			KeyPEM:   []byte("key-data"),
			SecretID: "secret-1",
		},
	}

	if err := manager.UpdateCertificates(initialCerts); err != nil {
		t.Fatalf("UpdateCertificates failed: %v", err)
	}

	// Get certificates and modify them
	certs := manager.GetCertificates()
	certs["example.com"].SecretID = "modified-secret"

	// Get certificates again and verify original is unchanged
	certsAgain := manager.GetCertificates()
	if certsAgain["example.com"].SecretID == "modified-secret" {
		t.Error("GetCertificates returned a mutable reference - modification affected internal state")
	}
}

func TestTLSManagerGetTLSConfig(t *testing.T) {
	logger := zap.NewNop()

	manager := NewTLSManager(logger)

	// Test that GetTLSConfig returns a valid config
	tlsConfig := manager.GetTLSConfig()
	if tlsConfig == nil {
		t.Fatal("GetTLSConfig returned nil")
	}

	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion TLS12, got %v", tlsConfig.MinVersion)
	}

	if tlsConfig.GetCertificate == nil {
		t.Error("GetCertificate callback not set")
	}
}
