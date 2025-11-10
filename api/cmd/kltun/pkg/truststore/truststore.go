package truststore

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// CAUniqueName is the unique identifier for our CA in trust stores
	CAUniqueName = "kloudlite-ca"
)

var (
	// sudoWarningOnce ensures we only show sudo warning once
	sudoWarningOnce sync.Once
)

// TrustStore represents a platform-specific trust store
type TrustStore interface {
	// Install installs the CA certificate to this trust store
	Install(certPath string, cert *x509.Certificate) error

	// Uninstall removes the CA certificate from this trust store
	Uninstall(cert *x509.Certificate) error

	// IsInstalled checks if the CA is already installed
	IsInstalled(cert *x509.Certificate) bool

	// Name returns the name of this trust store
	Name() string
}

// InstallAll installs the CA certificate to all available trust stores
func InstallAll(certPath string, trustStores []string) error {
	cert, err := LoadCertificate(certPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	stores := getEnabledStores(trustStores)

	var errors []string
	for _, store := range stores {
		log.Printf("Installing CA certificate to %s trust store...", store.Name())

		if store.IsInstalled(cert) {
			log.Printf("  ✓ Already installed in %s", store.Name())
			continue
		}

		if err := store.Install(certPath, cert); err != nil {
			errMsg := fmt.Sprintf("failed to install to %s: %v", store.Name(), err)
			log.Printf("  ✗ %s", errMsg)
			errors = append(errors, errMsg)
		} else {
			log.Printf("  ✓ Successfully installed to %s", store.Name())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some installations failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// UninstallAll removes the CA certificate from all available trust stores
func UninstallAll(certPath string, trustStores []string) error {
	cert, err := LoadCertificate(certPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	stores := getEnabledStores(trustStores)

	var errors []string
	for _, store := range stores {
		log.Printf("Uninstalling CA certificate from %s trust store...", store.Name())

		if !store.IsInstalled(cert) {
			log.Printf("  ✓ Not installed in %s", store.Name())
			continue
		}

		if err := store.Uninstall(cert); err != nil {
			errMsg := fmt.Sprintf("failed to uninstall from %s: %v", store.Name(), err)
			log.Printf("  ✗ %s", errMsg)
			errors = append(errors, errMsg)
		} else {
			log.Printf("  ✓ Successfully uninstalled from %s", store.Name())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some uninstallations failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// getEnabledStores returns the list of enabled trust stores
func getEnabledStores(trustStores []string) []TrustStore {
	// If no specific stores requested, use all available
	if len(trustStores) == 0 {
		trustStores = []string{"system", "nss", "java"}
	}

	var stores []TrustStore

	for _, name := range trustStores {
		switch strings.ToLower(name) {
		case "system":
			stores = append(stores, NewSystemStore())
		case "nss", "firefox":
			if store := NewNSSStore(); store != nil {
				stores = append(stores, store)
			}
		case "java":
			if store := NewJavaStore(); store != nil {
				stores = append(stores, store)
			}
		}
	}

	return stores
}

// LoadCertificate loads and parses a PEM-encoded certificate
func LoadCertificate(certPath string) (*x509.Certificate, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// CommandWithSudo wraps a command with sudo if not running as root
func CommandWithSudo(cmd ...string) *exec.Cmd {
	// Check if already running as root
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		return exec.Command(cmd[0], cmd[1:]...)
	}

	// Check if sudo is available
	if !BinaryExists("sudo") {
		sudoWarningOnce.Do(func() {
			log.Println(`Warning: "sudo" is not available, but is needed for system-level operations.`)
			log.Println(`Install sudo with your system package manager and try again.`)
		})
		return exec.Command(cmd[0], cmd[1:]...)
	}

	// Use sudo with custom prompt
	return exec.Command("sudo", append([]string{
		"--prompt=Sudo password:", "--",
	}, cmd...)...)
}

// BinaryExists checks if a binary exists in PATH
func BinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExecCommand executes a command and returns stdout, stderr, and error
func ExecCommand(cmd *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Combine stdout and stderr for better error messages
		combined := append(stdout.Bytes(), stderr.Bytes()...)
		return combined, err
	}

	return stdout.Bytes(), nil
}

// ExpandHomeDir expands ~ to the user's home directory
func ExpandHomeDir(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	u, err := user.Current()
	if err != nil {
		return path
	}

	return filepath.Join(u.HomeDir, path[1:])
}

// FindFiles finds files matching a glob pattern
func FindFiles(pattern string) ([]string, error) {
	pattern = ExpandHomeDir(pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter out non-regular files
	var files []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && info.Mode().IsRegular() {
			files = append(files, match)
		}
	}

	return files, nil
}

// FindDirs finds directories matching a glob pattern
func FindDirs(pattern string) ([]string, error) {
	pattern = ExpandHomeDir(pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter out non-directories
	var dirs []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && info.IsDir() {
			dirs = append(dirs, match)
		}
	}

	return dirs, nil
}

// GetCertFingerprint returns the SHA256 fingerprint of a certificate
func GetCertFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%X", cert.Signature)
}
