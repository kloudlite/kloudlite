//go:build darwin

package truststore

import (
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	// Step 1: Add certificate to system keychain
	cmd := CommandWithSudo("security", "add-trusted-cert", "-d",
		"-k", "/Library/Keychains/System.keychain", certPath)

	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to add certificate to keychain: %w\nOutput: %s", err, out)
	}

	// Step 2: Export trust settings
	tmpPlist, err := os.CreateTemp("", "kloudlite-trust-*.plist")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpPlist.Name()
	tmpPlist.Close()
	defer os.Remove(tmpPath)

	cmd = CommandWithSudo("security", "trust-settings-export", "-d", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to export trust settings: %w\nOutput: %s", err, out)
	}

	// Step 3: Modify trust settings
	if err := s.modifyTrustSettings(tmpPath, cert); err != nil {
		return fmt.Errorf("failed to modify trust settings: %w", err)
	}

	// Step 4: Import modified trust settings
	cmd = CommandWithSudo("security", "trust-settings-import", "-d", tmpPath)
	if out, err := ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to import trust settings: %w\nOutput: %s", err, out)
	}

	return nil
}

func (s *macOSStore) Uninstall(cert *x509.Certificate) error {
	// Remove the certificate from system keychain
	// We need to find the certificate file first
	cmd := exec.Command("security", "find-certificate", "-c", cert.Subject.CommonName,
		"-p", "/Library/Keychains/System.keychain")

	out, err := ExecCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to find certificate: %w", err)
	}

	// Create a temporary file with the certificate
	tmpCert, err := os.CreateTemp("", "kloudlite-cert-*.pem")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpCert.Name()
	tmpCert.Write(out)
	tmpCert.Close()
	defer os.Remove(tmpPath)

	// Remove the certificate
	cmd = CommandWithSudo("security", "remove-trusted-cert", "-d", tmpPath)
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
			"kSecTrustSettingsResult":       uint64(1),                                                      // kSecTrustSettingsResultTrustRoot
			"kSecTrustSettingsPolicyString": "sslServer",
		},
		{
			"kSecTrustSettingsPolicy": []byte{0x2A, 0x86, 0x48, 0x86, 0xF7, 0x63, 0x64, 0x01, 0x00}, // basicX509
			"kSecTrustSettingsResult": uint64(1),                                                      // kSecTrustSettingsResultTrustRoot
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
