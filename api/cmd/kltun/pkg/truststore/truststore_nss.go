package truststore

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// nssStore implements TrustStore for NSS/Firefox certificate databases
type nssStore struct {
	certutilPath string
	profiles     []string
}

// NewNSSStore creates a new NSS/Firefox trust store
func NewNSSStore() TrustStore {
	// Find certutil
	certutilPath := findCertutil()
	if certutilPath == "" {
		log.Println("Warning: certutil not found. NSS/Firefox trust store will be skipped.")
		log.Println(getNSSInstallGuide())
		return nil
	}

	// Find NSS profiles
	profiles := findNSSProfiles()
	if len(profiles) == 0 {
		log.Println("Warning: no NSS certificate databases found.")
		return nil
	}

	return &nssStore{
		certutilPath: certutilPath,
		profiles:     profiles,
	}
}

func (s *nssStore) Name() string {
	return "NSS/Firefox"
}

func (s *nssStore) IsInstalled(cert *x509.Certificate) bool {
	// Check if installed in any profile
	for _, profile := range s.profiles {
		dbPrefix := getNSSDBPrefix(profile)
		cmd := exec.Command(s.certutilPath, "-V",
			"-d", dbPrefix+profile,
			"-u", "L",
			"-n", CAUniqueName)

		if err := cmd.Run(); err == nil {
			return true
		}
	}

	return false
}

func (s *nssStore) Install(certPath string, cert *x509.Certificate) error {
	var errors []string

	for _, profile := range s.profiles {
		dbPrefix := getNSSDBPrefix(profile)

		cmd := exec.Command(s.certutilPath, "-A",
			"-d", dbPrefix+profile,
			"-t", "C,,",
			"-n", CAUniqueName,
			"-i", certPath)

		if err := s.execCertutil(cmd); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", profile, err))
		} else {
			log.Printf("    ✓ Installed to %s", profile)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to install to some profiles:\n    %s", strings.Join(errors, "\n    "))
	}

	return nil
}

func (s *nssStore) Uninstall(cert *x509.Certificate) error {
	var errors []string

	for _, profile := range s.profiles {
		dbPrefix := getNSSDBPrefix(profile)

		cmd := exec.Command(s.certutilPath, "-D",
			"-d", dbPrefix+profile,
			"-n", CAUniqueName)

		if err := s.execCertutil(cmd); err != nil {
			// Ignore errors if certificate doesn't exist
			if !strings.Contains(err.Error(), "not found") {
				errors = append(errors, fmt.Sprintf("%s: %v", profile, err))
			}
		} else {
			log.Printf("    ✓ Uninstalled from %s", profile)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to uninstall from some profiles:\n    %s", strings.Join(errors, "\n    "))
	}

	return nil
}

// execCertutil executes certutil with automatic sudo retry on permission errors
func (s *nssStore) execCertutil(cmd *exec.Cmd) error {
	out, err := ExecCommand(cmd)
	if err != nil {
		// Check for permission errors
		outStr := string(out)
		if strings.Contains(outStr, "SEC_ERROR_READ_ONLY") {
			// Retry with sudo
			sudoCmd := CommandWithSudo(cmd.Args...)
			out, err = ExecCommand(sudoCmd)
			if err != nil {
				return fmt.Errorf("%w\nOutput: %s", err, out)
			}
		} else {
			return fmt.Errorf("%w\nOutput: %s", err, out)
		}
	}

	return nil
}

// findCertutil finds the certutil binary
func findCertutil() string {
	// Check in PATH first
	if path, err := exec.LookPath("certutil"); err == nil {
		return path
	}

	// macOS-specific paths
	if runtime.GOOS == "darwin" {
		// Check Homebrew paths
		brewPaths := []string{
			"/usr/local/opt/nss/bin/certutil",
			"/opt/homebrew/opt/nss/bin/certutil",
		}

		for _, path := range brewPaths {
			if PathExists(path) {
				return path
			}
		}

		// Try to get from brew
		cmd := exec.Command("brew", "--prefix", "nss")
		if out, err := ExecCommand(cmd); err == nil {
			brewPrefix := strings.TrimSpace(string(out))
			certutilPath := filepath.Join(brewPrefix, "bin", "certutil")
			if PathExists(certutilPath) {
				return certutilPath
			}
		}
	}

	return ""
}

// findNSSProfiles finds all NSS certificate database profiles
func findNSSProfiles() []string {
	var profiles []string

	// Standard NSS databases
	nssPaths := []string{
		"~/.pki/nssdb",
		"/etc/pki/nssdb", // CentOS 7 system-wide
	}

	// Snap-packaged browsers
	if runtime.GOOS == "linux" {
		nssPaths = append(nssPaths, "~/snap/chromium/current/.pki/nssdb")
	}

	for _, path := range nssPaths {
		path = ExpandHomeDir(path)
		if hasNSSDB(path) {
			profiles = append(profiles, path)
		}
	}

	// Firefox profiles
	firefoxProfiles := findFirefoxProfiles()
	profiles = append(profiles, firefoxProfiles...)

	return profiles
}

// findFirefoxProfiles finds Firefox profile directories
func findFirefoxProfiles() []string {
	var profiles []string

	var firefoxDirs []string
	switch runtime.GOOS {
	case "darwin":
		firefoxDirs = []string{
			"~/Library/Application Support/Firefox/Profiles/*",
		}
	case "linux":
		firefoxDirs = []string{
			"~/.mozilla/firefox/*",
			"~/snap/firefox/common/.mozilla/firefox/*",
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			firefoxDirs = []string{
				filepath.Join(appData, "Mozilla", "Firefox", "Profiles", "*"),
			}
		}
	}

	for _, pattern := range firefoxDirs {
		dirs, _ := FindDirs(pattern)
		for _, dir := range dirs {
			if hasNSSDB(dir) {
				profiles = append(profiles, dir)
			}
		}
	}

	return profiles
}

// hasNSSDB checks if a directory contains an NSS certificate database
func hasNSSDB(path string) bool {
	// Check for SQL format (modern)
	if PathExists(filepath.Join(path, "cert9.db")) {
		return true
	}

	// Check for DBM format (legacy)
	if PathExists(filepath.Join(path, "cert8.db")) {
		return true
	}

	return false
}

// getNSSDBPrefix returns the appropriate database prefix (sql: or dbm:)
func getNSSDBPrefix(path string) string {
	if PathExists(filepath.Join(path, "cert9.db")) {
		return "sql:"
	}
	if PathExists(filepath.Join(path, "cert8.db")) {
		return "dbm:"
	}
	return "sql:" // default to SQL format
}

// getNSSInstallGuide returns platform-specific installation instructions
func getNSSInstallGuide() string {
	switch runtime.GOOS {
	case "darwin":
		return "  Install NSS tools with: brew install nss"
	case "linux":
		guides := []string{
			"  Debian/Ubuntu: apt install libnss3-tools",
			"  Red Hat/CentOS: yum install nss-tools",
			"  Arch Linux: pacman -S nss",
			"  openSUSE: zypper install mozilla-nss-tools",
		}
		return strings.Join(guides, "\n")
	case "windows":
		return "  NSS tools are typically bundled with Firefox on Windows"
	default:
		return "  Install NSS tools with your system package manager"
	}
}
