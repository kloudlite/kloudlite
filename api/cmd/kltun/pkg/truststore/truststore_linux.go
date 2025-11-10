//go:build linux

package truststore

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// linuxStore implements TrustStore for Linux system certificate stores
type linuxStore struct {
	distro       string
	certPath     string
	certExt      string
	updateCmd    string
	updateArgs   []string
	legacyPaths  []string
	installGuide string
}

// NewSystemStore creates a new Linux system trust store
func NewSystemStore() TrustStore {
	// Detect distribution by checking for specific paths
	// Order matters - check more specific paths first

	if PathExists("/etc/pki/ca-trust/source/anchors/") {
		// Red Hat/Fedora/CentOS
		return &linuxStore{
			distro:      "Red Hat/Fedora/CentOS",
			certPath:    "/etc/pki/ca-trust/source/anchors",
			certExt:     ".pem",
			updateCmd:   "update-ca-trust",
			updateArgs:  []string{"extract"},
			installGuide: "Install NSS tools with: yum install nss-tools",
		}
	}

	if PathExists("/usr/local/share/ca-certificates/") {
		// Debian/Ubuntu
		return &linuxStore{
			distro:      "Debian/Ubuntu",
			certPath:    "/usr/local/share/ca-certificates",
			certExt:     ".crt",
			updateCmd:   "update-ca-certificates",
			updateArgs:  []string{},
			installGuide: "Install NSS tools with: apt install libnss3-tools",
		}
	}

	if PathExists("/etc/ca-certificates/trust-source/anchors/") {
		// Arch Linux
		return &linuxStore{
			distro:      "Arch Linux",
			certPath:    "/etc/ca-certificates/trust-source/anchors",
			certExt:     ".crt",
			updateCmd:   "trust",
			updateArgs:  []string{"extract-compat"},
			installGuide: "Install NSS tools with: pacman -S nss",
		}
	}

	if PathExists("/usr/share/pki/trust/anchors") {
		// openSUSE
		return &linuxStore{
			distro:      "openSUSE",
			certPath:    "/usr/share/pki/trust/anchors",
			certExt:     ".pem",
			updateCmd:   "update-ca-certificates",
			updateArgs:  []string{},
			installGuide: "Install NSS tools with: zypper install mozilla-nss-tools",
		}
	}

	log.Println("Warning: unsupported Linux distribution, using generic approach")
	return &linuxStore{
		distro:      "Generic Linux",
		certPath:    "/usr/local/share/ca-certificates",
		certExt:     ".crt",
		updateCmd:   "update-ca-certificates",
		updateArgs:  []string{},
		installGuide: "Install NSS tools with your package manager",
	}
}

func (s *linuxStore) Name() string {
	return fmt.Sprintf("Linux System (%s)", s.distro)
}

func (s *linuxStore) IsInstalled(cert *x509.Certificate) bool {
	certFile := filepath.Join(s.certPath, CAUniqueName+s.certExt)
	return PathExists(certFile)
}

func (s *linuxStore) Install(certPath string, cert *x509.Certificate) error {
	// Read certificate file
	data, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	// Determine target path
	targetPath := filepath.Join(s.certPath, CAUniqueName+s.certExt)

	// Write certificate to system trust store using sudo
	// We use a pipe to avoid creating temp files
	cmd := CommandWithSudo("tee", targetPath)
	cmd.Stdin = os.NewFile(0, "/dev/stdin")

	// Create a pipe for stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tee command: %w", err)
	}

	// Write certificate data
	if _, err := stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}
	stdin.Close()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to write certificate to %s: %w", targetPath, err)
	}

	// Update system trust store
	updateCmd := CommandWithSudo(s.updateCmd, s.updateArgs...)
	if out, err := ExecCommand(updateCmd); err != nil {
		return fmt.Errorf("failed to update trust store: %w\nOutput: %s", err, out)
	}

	log.Printf("  Note: %s", s.installGuide)

	return nil
}

func (s *linuxStore) Uninstall(cert *x509.Certificate) error {
	// Remove certificate file
	certFile := filepath.Join(s.certPath, CAUniqueName+s.certExt)

	// Also remove legacy filenames (in case they exist from older versions)
	filesToRemove := []string{
		certFile,
		filepath.Join(s.certPath, CAUniqueName+".pem"),
		filepath.Join(s.certPath, CAUniqueName+".crt"),
		filepath.Join(s.certPath, "kloudlite.pem"),
		filepath.Join(s.certPath, "kloudlite.crt"),
	}

	for _, file := range filesToRemove {
		cmd := CommandWithSudo("rm", "-f", file)
		// Ignore errors since files might not exist
		ExecCommand(cmd)
	}

	// Update system trust store
	updateCmd := CommandWithSudo(s.updateCmd, s.updateArgs...)
	if out, err := ExecCommand(updateCmd); err != nil {
		return fmt.Errorf("failed to update trust store: %w\nOutput: %s", err, out)
	}

	return nil
}
