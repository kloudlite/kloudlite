//go:build darwin

package truststore

import (
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"howett.net/plist"
)

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
	// Try to find the certificate in the system keychain
	cmd := exec.Command("security", "find-certificate", "-c", cert.Subject.CommonName,
		"-a", "-Z", "/Library/Keychains/System.keychain")

	out, err := ExecCommand(cmd)
	if err != nil {
		return false
	}

	// Check if our certificate is in the output
	return strings.Contains(string(out), cert.Subject.CommonName)
}

func (s *macOSStore) Install(certPath string, cert *x509.Certificate) error {
	// The daemon runs as root, so we can add the certificate as a trusted root.
	// Use -r trustRoot to set it as a trusted root certificate.
	// Avoid -d (admin trust settings domain) which requires GUI authorization.

	// Check if already running as root (daemon case)
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		// Running as root - use add-trusted-cert with -r trustRoot
		// This adds the cert and sets trust settings in one command
		cmd := exec.Command("security", "add-trusted-cert",
			"-r", "trustRoot",
			"-k", "/Library/Keychains/System.keychain",
			certPath)
		if out, err := ExecCommand(cmd); err != nil {
			return fmt.Errorf("failed to add trusted certificate: %w\nOutput: %s", err, out)
		}
		return nil
	}

	// Not running as root - this shouldn't happen if called from daemon
	// Fall back to using sudo
	cmd := CommandWithSudo("security", "add-trusted-cert",
		"-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain",
		certPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to add trusted certificate: %w\nOutput: %s", err, out)
	}

	return nil
}

// setTrustSettingsAsRoot sets trust settings when running as root
// This uses the plist method which works without GUI interaction
func (s *macOSStore) setTrustSettingsAsRoot(certPath string, cert *x509.Certificate) error {
	tmpPlist, err := os.CreateTemp("", "kloudlite-trust-*.plist")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpPlist.Name()
	tmpPlist.Close()
	defer os.Remove(tmpPath)

	// Export current trust settings
	cmd := exec.Command("security", "trust-settings-export", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		// If no trust settings exist yet, create empty plist
		log.Printf("No existing trust settings, will create new: %v", err)
		emptyPlist := map[string]interface{}{
			"trustVersion": uint64(1),
			"trustList":    make(map[string]interface{}),
		}
		data, _ := plist.MarshalIndent(emptyPlist, plist.XMLFormat, "\t")
		if err := os.WriteFile(tmpPath, data, 0600); err != nil {
			return fmt.Errorf("failed to create empty trust settings: %w", err)
		}
	} else {
		_ = out
	}

	// Modify trust settings to trust our CA
	if err := s.modifyTrustSettings(tmpPath, cert); err != nil {
		return fmt.Errorf("failed to modify trust settings: %w", err)
	}

	// Import modified trust settings
	cmd = exec.Command("security", "trust-settings-import", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to import trust settings: %w\nOutput: %s", err, out)
	}

	return nil
}

// setTrustSettingsWithSudo sets trust settings using sudo
func (s *macOSStore) setTrustSettingsWithSudo(certPath string, cert *x509.Certificate) error {
	tmpPlist, err := os.CreateTemp("", "kloudlite-trust-*.plist")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpPlist.Name()
	tmpPlist.Close()
	defer os.Remove(tmpPath)

	cmd := CommandWithSudo("security", "trust-settings-export", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		log.Printf("No existing trust settings, will create new: %v", err)
		emptyPlist := map[string]interface{}{
			"trustVersion": uint64(1),
			"trustList":    make(map[string]interface{}),
		}
		data, _ := plist.MarshalIndent(emptyPlist, plist.XMLFormat, "\t")
		if err := os.WriteFile(tmpPath, data, 0600); err != nil {
			return fmt.Errorf("failed to create empty trust settings: %w", err)
		}
	} else {
		_ = out
	}

	if err := s.modifyTrustSettings(tmpPath, cert); err != nil {
		return fmt.Errorf("failed to modify trust settings: %w", err)
	}

	cmd = CommandWithSudo("security", "trust-settings-import", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to import trust settings: %w\nOutput: %s", err, out)
	}

	return nil
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

// modifyTrustSettings modifies the trust settings plist to explicitly trust our CA
func (s *macOSStore) modifyTrustSettings(plistPath string, cert *x509.Certificate) error {
	// Read the plist file
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return fmt.Errorf("failed to read plist: %w", err)
	}

	// Parse the plist
	var trustSettings map[string]interface{}
	if _, err := plist.Unmarshal(data, &trustSettings); err != nil {
		return fmt.Errorf("failed to parse plist: %w", err)
	}

	// Validate trust version
	if version, ok := trustSettings["trustVersion"].(uint64); !ok || version != 1 {
		return fmt.Errorf("unsupported trust settings version")
	}

	// Get trust list
	trustList, ok := trustSettings["trustList"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid trust list format")
	}

	// Encode certificate subject for matching
	subjectASN1, err := asn1.Marshal(cert.Subject.ToRDNSequence())
	if err != nil {
		return fmt.Errorf("failed to encode subject: %w", err)
	}

	// Find our certificate entry
	subjectKey := fmt.Sprintf("%X", subjectASN1)
	var certEntry map[string]interface{}

	for key, value := range trustList {
		if strings.EqualFold(key, subjectKey) {
			certEntry = value.(map[string]interface{})
			break
		}
	}

	if certEntry == nil {
		log.Printf("Warning: certificate not found in trust list, creating new entry")
		certEntry = make(map[string]interface{})
		trustList[subjectKey] = certEntry
	}

	// Set trust settings for SSL and X.509
	certTrustSettings := []map[string]interface{}{
		{
			"kSecTrustSettingsPolicy":       []byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x63, 0x64, 0x01, 0x01}, // sslServer
			"kSecTrustSettingsResult":       uint64(1),                                                    // kSecTrustSettingsResultTrustRoot
			"kSecTrustSettingsPolicyString": "sslServer",
		},
		{
			"kSecTrustSettingsPolicy": []byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x63, 0x64, 0x01, 0x00}, // basicX509
			"kSecTrustSettingsResult": uint64(1),                                                    // kSecTrustSettingsResultTrustRoot
		},
	}

	certEntry["trustSettings"] = certTrustSettings

	// Write back the modified plist
	data, err = plist.MarshalIndent(trustSettings, plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal plist: %w", err)
	}

	if err := os.WriteFile(plistPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	return nil
}
