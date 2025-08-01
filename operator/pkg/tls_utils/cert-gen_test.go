package tls_utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	fn "github.com/kloudlite/operator/pkg/functions"
)

func TestGenTLSCert(t *testing.T) {
	tests := []struct {
		name    string
		args    GenTLSCertArgs
		wantErr bool
		checks  func(t *testing.T, caBundle, tlsCert, tlsKey []byte)
	}{
		{
			name: "valid certificate generation with defaults",
			args: GenTLSCertArgs{
				DNSNames: []string{"example.com", "*.example.com"},
			},
			wantErr: false,
			checks: func(t *testing.T, caBundle, tlsCert, tlsKey []byte) {
				// Verify all outputs are non-empty
				if len(caBundle) == 0 {
					t.Error("CA bundle is empty")
				}
				if len(tlsCert) == 0 {
					t.Error("TLS cert is empty")
				}
				if len(tlsKey) == 0 {
					t.Error("TLS key is empty")
				}
			},
		},
		{
			name: "custom time range",
			args: GenTLSCertArgs{
				DNSNames:  []string{"test.local"},
				NotBefore: fn.New(time.Now().Add(-24 * time.Hour)),
				NotAfter:  fn.New(time.Now().Add(365 * 24 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name: "custom certificate label",
			args: GenTLSCertArgs{
				DNSNames:         []string{"api.service.local"},
				CertificateLabel: "API Service",
			},
			wantErr: false,
		},
		{
			name: "no DNS names provided",
			args: GenTLSCertArgs{
				DNSNames: []string{},
			},
			wantErr: true,
		},
		{
			name: "multiple DNS names",
			args: GenTLSCertArgs{
				DNSNames: []string{"service1.local", "service2.local", "service3.local"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caBundle, tlsCert, tlsKey, err := GenTLSCert(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GenTLSCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Run custom checks if provided
			if tt.checks != nil {
				tt.checks(t, caBundle, tlsCert, tlsKey)
			}

			// Validate CA certificate
			t.Run("validate CA certificate", func(t *testing.T) {
				validateCACertificate(t, caBundle)
			})

			// Validate server certificate
			t.Run("validate server certificate", func(t *testing.T) {
				validateServerCertificate(t, tlsCert, tlsKey, tt.args.DNSNames)
			})

			// Validate certificate chain
			t.Run("validate certificate chain", func(t *testing.T) {
				validateCertificateChain(t, caBundle, tlsCert)
			})

			// Validate TLS configuration
			t.Run("validate TLS configuration", func(t *testing.T) {
				validateTLSConfig(t, tlsCert, tlsKey)
			})
		})
	}
}

func validateCACertificate(t *testing.T, caCertPEM []byte) {
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		t.Fatal("Failed to decode CA certificate PEM")
	}

	if block.Type != "CERTIFICATE" {
		t.Errorf("Expected PEM type CERTIFICATE, got %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	// Verify CA properties
	if !cert.IsCA {
		t.Error("Certificate is not marked as CA")
	}

	if cert.Subject.Organization[0] != "Kloudlite CA" {
		t.Errorf("Expected organization 'Kloudlite CA', got %v", cert.Subject.Organization)
	}

	// Check key usage
	expectedKeyUsage := x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	if cert.KeyUsage != expectedKeyUsage {
		t.Errorf("Expected key usage %v, got %v", expectedKeyUsage, cert.KeyUsage)
	}

	// Verify validity period
	if cert.NotBefore.After(time.Now()) {
		t.Error("CA certificate NotBefore is in the future")
	}

	if cert.NotAfter.Before(time.Now()) {
		t.Error("CA certificate has expired")
	}
}

func validateServerCertificate(t *testing.T, certPEM, keyPEM []byte, expectedDNSNames []string) {
	// Parse certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("Failed to decode server certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse server certificate: %v", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		t.Fatal("Failed to decode private key PEM")
	}

	if keyBlock.Type != "EC PRIVATE KEY" {
		t.Errorf("Expected PEM type 'EC PRIVATE KEY', got %s", keyBlock.Type)
	}

	// Verify certificate properties
	if cert.IsCA {
		t.Error("Server certificate should not be a CA")
	}

	// Check DNS names
	if len(cert.DNSNames) != len(expectedDNSNames) {
		t.Errorf("Expected %d DNS names, got %d", len(expectedDNSNames), len(cert.DNSNames))
	}

	for i, name := range expectedDNSNames {
		if i >= len(cert.DNSNames) || cert.DNSNames[i] != name {
			t.Errorf("Expected DNS name %s at position %d", name, i)
		}
	}

	// Check key usage
	expectedKeyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if cert.KeyUsage != expectedKeyUsage {
		t.Errorf("Expected key usage %v, got %v", expectedKeyUsage, cert.KeyUsage)
	}

	// Check extended key usage
	if len(cert.ExtKeyUsage) != 1 || cert.ExtKeyUsage[0] != x509.ExtKeyUsageServerAuth {
		t.Error("Expected ExtKeyUsage to contain only ServerAuth")
	}
}

func validateCertificateChain(t *testing.T, caCertPEM, serverCertPEM []byte) {
	// Parse CA certificate
	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	// Parse server certificate
	serverBlock, _ := pem.Decode(serverCertPEM)
	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse server certificate: %v", err)
	}

	// Create CA pool
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	// Verify server certificate against CA
	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if _, err := serverCert.Verify(opts); err != nil {
		t.Errorf("Failed to verify server certificate against CA: %v", err)
	}

	// Verify that CA and server cert have same validity period
	if !caCert.NotAfter.Equal(serverCert.NotAfter) {
		t.Errorf("CA and server certificate have different NotAfter times: CA=%v, Server=%v", 
			caCert.NotAfter, serverCert.NotAfter)
	}
}

func validateTLSConfig(t *testing.T, certPEM, keyPEM []byte) {
	// Test that the certificate and key can be loaded as a TLS certificate
	_, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Errorf("Failed to load certificate and key as TLS pair: %v", err)
	}
}

func TestGenTLSCertTimeValidation(t *testing.T) {
	// Test that certificates are valid for the specified time range
	// x509 certificates store times in UTC with second precision
	now := time.Now().UTC().Truncate(time.Second)
	notBefore := now.Add(-time.Hour)
	notAfter := now.Add(24 * time.Hour)

	args := GenTLSCertArgs{
		DNSNames:  []string{"time-test.local"},
		NotBefore: &notBefore,
		NotAfter:  &notAfter,
	}

	caBundle, tlsCert, _, err := GenTLSCert(args)
	if err != nil {
		t.Fatalf("Failed to generate certificates: %v", err)
	}

	// Parse and check CA certificate times
	caBlock, _ := pem.Decode(caBundle)
	caCert, _ := x509.ParseCertificate(caBlock.Bytes)
	
	// Compare times with truncation to account for certificate time precision
	if !caCert.NotBefore.Equal(notBefore.UTC().Truncate(time.Second)) {
		t.Errorf("CA NotBefore mismatch: expected %v, got %v", notBefore, caCert.NotBefore)
	}
	
	if !caCert.NotAfter.Equal(notAfter.UTC().Truncate(time.Second)) {
		t.Errorf("CA NotAfter mismatch: expected %v, got %v", notAfter, caCert.NotAfter)
	}

	// Parse and check server certificate times
	serverBlock, _ := pem.Decode(tlsCert)
	serverCert, _ := x509.ParseCertificate(serverBlock.Bytes)
	
	if !serverCert.NotBefore.Equal(notBefore.UTC().Truncate(time.Second)) {
		t.Errorf("Server NotBefore mismatch: expected %v, got %v", notBefore, serverCert.NotBefore)
	}
	
	if !serverCert.NotAfter.Equal(notAfter.UTC().Truncate(time.Second)) {
		t.Errorf("Server NotAfter mismatch: expected %v, got %v", notAfter, serverCert.NotAfter)
	}
}