//go:build linux

package truststore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a self-signed test certificate
func generateTestCert(t *testing.T, commonName string) (*x509.Certificate, []byte) {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return cert, certPEM
}

func TestLinuxStore_IsInstalled_NoCertFile(t *testing.T) {
	// Create a temp directory
	tmpDir := t.TempDir()

	store := &linuxStore{
		distro:   "Test",
		certPath: tmpDir,
		certExt:  ".crt",
	}

	cert, _ := generateTestCert(t, "Test CA")

	// Should return false when no cert file exists
	if store.IsInstalled(cert) {
		t.Error("IsInstalled should return false when cert file does not exist")
	}
}

func TestLinuxStore_IsInstalled_MatchingCert(t *testing.T) {
	tmpDir := t.TempDir()

	store := &linuxStore{
		distro:   "Test",
		certPath: tmpDir,
		certExt:  ".crt",
	}

	cert, certPEM := generateTestCert(t, "Test CA")

	// Write the cert to the expected location
	certFile := filepath.Join(tmpDir, CAUniqueName+".crt")
	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	// Should return true when same cert is installed
	if !store.IsInstalled(cert) {
		t.Error("IsInstalled should return true when matching cert is installed")
	}
}

func TestLinuxStore_IsInstalled_DifferentCert(t *testing.T) {
	tmpDir := t.TempDir()

	store := &linuxStore{
		distro:   "Test",
		certPath: tmpDir,
		certExt:  ".crt",
	}

	// Generate two different certs
	cert1, certPEM1 := generateTestCert(t, "Test CA 1")
	cert2, _ := generateTestCert(t, "Test CA 2")

	// Write cert1 to file
	certFile := filepath.Join(tmpDir, CAUniqueName+".crt")
	if err := os.WriteFile(certFile, certPEM1, 0644); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	// Check with cert1 - should return true
	if !store.IsInstalled(cert1) {
		t.Error("IsInstalled should return true for matching cert")
	}

	// Check with cert2 - should return false (different cert)
	if store.IsInstalled(cert2) {
		t.Error("IsInstalled should return false when different cert is installed")
	}
}

func TestLinuxStore_IsInstalled_CorruptedCertFile(t *testing.T) {
	tmpDir := t.TempDir()

	store := &linuxStore{
		distro:   "Test",
		certPath: tmpDir,
		certExt:  ".crt",
	}

	cert, _ := generateTestCert(t, "Test CA")

	// Write corrupted data to cert file
	certFile := filepath.Join(tmpDir, CAUniqueName+".crt")
	if err := os.WriteFile(certFile, []byte("not a valid certificate"), 0644); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	// Should return false when cert file is corrupted (triggers reinstall)
	if store.IsInstalled(cert) {
		t.Error("IsInstalled should return false when cert file is corrupted")
	}
}

func TestLinuxStore_IsInstalled_SameSubjectDifferentSerial(t *testing.T) {
	tmpDir := t.TempDir()

	store := &linuxStore{
		distro:   "Test",
		certPath: tmpDir,
		certExt:  ".crt",
	}

	// Generate two certs with same common name but different serial (simulates CA rotation)
	cert1, certPEM1 := generateTestCert(t, "Kloudlite CA for *.test.dev")
	// Sleep briefly to ensure different serial (based on UnixNano)
	time.Sleep(time.Millisecond)
	cert2, _ := generateTestCert(t, "Kloudlite CA for *.test.dev")

	// Verify they have same CN but different content
	if cert1.Subject.CommonName != cert2.Subject.CommonName {
		t.Fatal("test setup error: certs should have same CommonName")
	}
	if cert1.SerialNumber.Cmp(cert2.SerialNumber) == 0 {
		t.Fatal("test setup error: certs should have different serial numbers")
	}

	// Write cert1 to file
	certFile := filepath.Join(tmpDir, CAUniqueName+".crt")
	if err := os.WriteFile(certFile, certPEM1, 0644); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	// Check with cert2 - should return false (different cert, even with same CN)
	// This is the key test for CA rotation scenario
	if store.IsInstalled(cert2) {
		t.Error("IsInstalled should return false when cert with same CN but different serial is installed (CA rotation scenario)")
	}
}
