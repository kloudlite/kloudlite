//go:build darwin

package truststore

import (
	"crypto/x509"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
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
	// Use -d (admin trust settings domain) with -r trustRoot and -p ssl to set trust for SSL.
	// The -d flag works when running as root via sudo.

	// Check if already running as root (daemon case)
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		// Running as root - use add-trusted-cert with admin domain (-d)
		// -r trustRoot: set as trusted root certificate
		// -p ssl: trust for SSL policy (needed for browsers)
		cmd := exec.Command("security", "add-trusted-cert",
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

	// Not running as root - this shouldn't happen if called from daemon
	// Fall back to using sudo
	cmd := CommandWithSudo("security", "add-trusted-cert",
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
