//go:build darwin

package truststore

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

const (
	// adminTrustSettingsPath is where macOS stores admin trust settings
	adminTrustSettingsPath = "/var/db/TrustSettings/Admin.plist"
	trustSettingsDir       = "/var/db/TrustSettings"

	// Trust settings constants
	kSecTrustSettingsResultTrustRoot = 1 // Trust as root certificate

	// SSL policy OID (base64 of the OID for SSL)
	sslPolicyOID = "KoZIhvdjZAED" // kSecPolicyAppleSSL
)

// TrustSettings represents the structure of macOS trust settings plist
type TrustSettings struct {
	TrustList    map[string][]TrustSettingsEntry `plist:"trustList"`
	TrustVersion int                             `plist:"trustVersion"`
}

// TrustSettingsEntry represents a single trust setting entry
type TrustSettingsEntry struct {
	KSecTrustSettingsPolicy []byte `plist:"kSecTrustSettingsPolicy,omitempty"`
	KSecTrustSettingsResult int    `plist:"kSecTrustSettingsResult"`
}

// macOSStore implements TrustStore for macOS Keychain
type macOSStore struct{}

// NewSystemStore creates a new macOS system trust store
func NewSystemStore() TrustStore {
	return &macOSStore{}
}

func (s *macOSStore) Name() string {
	return "macOS Keychain"
}

func (s *macOSStore) IsInstalled(cert *x509.Certificate) bool {
	// Check if certificate is in the system keychain
	cmd := exec.Command("security", "find-certificate", "-c", cert.Subject.CommonName,
		"-a", "-Z", "/Library/Keychains/System.keychain")

	out, err := ExecCommand(cmd)
	if err != nil {
		return false
	}

	// Check if certificate exists in keychain
	if !strings.Contains(string(out), cert.Subject.CommonName) {
		return false
	}

	// Also check if trust settings exist in Admin.plist
	// This is important because the cert may be in keychain but not trusted
	return s.isTrusted(cert)
}

// isTrusted checks if the certificate has trust settings in Admin.plist
func (s *macOSStore) isTrusted(cert *x509.Certificate) bool {
	data, err := os.ReadFile(adminTrustSettingsPath)
	if err != nil {
		// No trust settings file means not trusted
		return false
	}

	var settings TrustSettings
	if _, err := plist.Unmarshal(data, &settings); err != nil {
		return false
	}

	if settings.TrustList == nil {
		return false
	}

	// Calculate SHA-1 hash of the certificate
	hash := sha1.Sum(cert.Raw)
	certHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Check if this cert hash exists in trust list
	_, exists := settings.TrustList[certHash]
	return exists
}

func (s *macOSStore) Install(certPath string, cert *x509.Certificate) error {
	// The daemon runs as root, so we can add the certificate to the keychain
	// and directly modify the trust settings plist file.
	//
	// Why direct plist manipulation instead of `security add-trusted-cert -d`?
	// The -d flag (admin trust settings) calls SecTrustSettingsSetTrustSettings,
	// which requires GUI authorization even when running as root. This is a macOS
	// security feature - root != authorization for trust settings.
	// By directly writing to /var/db/TrustSettings/Admin.plist, we bypass this
	// restriction since we have root filesystem access.

	// Check if already running as root (daemon case)
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		// Step 1: Add certificate to system keychain (this works without authorization)
		cmd := exec.Command("security", "add-certificates",
			"-k", "/Library/Keychains/System.keychain",
			certPath)
		if out, err := ExecCommand(cmd); err != nil {
			// Ignore "already exists" / "already in" errors
			outStr := string(out)
			if !strings.Contains(outStr, "already exists") &&
				!strings.Contains(outStr, "already in") &&
				!strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "already in") {
				return fmt.Errorf("failed to add certificate to keychain: %w\nOutput: %s", err, out)
			}
		}

		// Step 2: Directly update the admin trust settings plist
		if err := s.updateAdminTrustSettings(cert); err != nil {
			return fmt.Errorf("failed to update trust settings: %w", err)
		}

		return nil
	}

	// Not running as root - this shouldn't happen if called from daemon
	// Fall back to using sudo for both operations
	cmd := CommandWithSudo("security", "add-certificates",
		"-k", "/Library/Keychains/System.keychain",
		certPath)
	if out, err := ExecCommand(cmd); err != nil {
		outStr := string(out)
		if !strings.Contains(outStr, "already exists") &&
			!strings.Contains(outStr, "already in") &&
			!strings.Contains(err.Error(), "already exists") &&
			!strings.Contains(err.Error(), "already in") {
			return fmt.Errorf("failed to add certificate to keychain: %w\nOutput: %s", err, out)
		}
	}

	// For non-root, we still need to use the security command with sudo
	// This will prompt for GUI authorization
	cmd = CommandWithSudo("security", "add-trusted-cert",
		"-d",
		"-r", "trustRoot",
		"-p", "ssl",
		"-k", "/Library/Keychains/System.keychain",
		certPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to add trusted certificate: %w\nOutput: %s", err, out)
	}

	return nil
}

// updateAdminTrustSettings directly modifies the Admin.plist trust settings file
func (s *macOSStore) updateAdminTrustSettings(cert *x509.Certificate) error {
	// Calculate SHA-1 hash of the certificate (used as key in trust settings)
	hash := sha1.Sum(cert.Raw)
	certHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Ensure the trust settings directory exists
	if err := os.MkdirAll(trustSettingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create trust settings directory: %w", err)
	}

	// Load existing trust settings or create new
	var settings TrustSettings
	data, err := os.ReadFile(adminTrustSettingsPath)
	if err == nil {
		// Parse existing plist
		if _, err := plist.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing trust settings: %w", err)
		}
	}

	// Initialize trust list if nil
	if settings.TrustList == nil {
		settings.TrustList = make(map[string][]TrustSettingsEntry)
	}
	settings.TrustVersion = 1

	// Decode the SSL policy OID
	sslPolicy, err := decodeBase64(sslPolicyOID)
	if err != nil {
		return fmt.Errorf("failed to decode SSL policy OID: %w", err)
	}

	// Add trust entry for SSL (kSecPolicyAppleSSL)
	settings.TrustList[certHash] = []TrustSettingsEntry{
		{
			KSecTrustSettingsPolicy: sslPolicy,
			KSecTrustSettingsResult: kSecTrustSettingsResultTrustRoot,
		},
	}

	// Marshal to plist
	plistData, err := plist.MarshalIndent(settings, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal trust settings: %w", err)
	}

	// Write atomically using temp file
	tmpFile := filepath.Join(trustSettingsDir, ".Admin.plist.tmp")
	if err := os.WriteFile(tmpFile, plistData, 0644); err != nil {
		return fmt.Errorf("failed to write temp trust settings file: %w", err)
	}

	if err := os.Rename(tmpFile, adminTrustSettingsPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename trust settings file: %w", err)
	}

	return nil
}

// decodeBase64 decodes a base64 string
func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func (s *macOSStore) Uninstall(cert *x509.Certificate) error {
	// Remove the certificate from system keychain
	// We need to find and delete the certificate by its common name

	// Check if already running as root (daemon case)
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		// Delete directly from system keychain - no -d flag needed when root
		cmd := exec.Command("security", "delete-certificate", "-c", cert.Subject.CommonName,
			"-t", "/Library/Keychains/System.keychain")
		if out, err := ExecCommand(cmd); err != nil {
			return fmt.Errorf("failed to remove certificate: %w\nOutput: %s", err, out)
		}
		return nil
	}

	// Not running as root - use sudo
	cmd := CommandWithSudo("security", "delete-certificate", "-c", cert.Subject.CommonName,
		"-t", "/Library/Keychains/System.keychain")
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to remove certificate: %w\nOutput: %s", err, out)
	}

	return nil
}
