package truststore

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// javaStore implements TrustStore for Java keystore
type javaStore struct {
	javaHome    string
	keytoolPath string
	cacertsPath string
	storePass   string
}

// NewJavaStore creates a new Java keystore trust store
func NewJavaStore() TrustStore {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		log.Println("Warning: JAVA_HOME not set. Java keystore will be skipped.")
		return nil
	}

	// Find keytool
	keytoolPath := findKeytool(javaHome)
	if keytoolPath == "" {
		log.Printf("Warning: keytool not found in JAVA_HOME (%s). Java keystore will be skipped.", javaHome)
		return nil
	}

	// Find cacerts
	cacertsPath := findCacerts(javaHome)
	if cacertsPath == "" {
		log.Printf("Warning: cacerts not found in JAVA_HOME (%s). Java keystore will be skipped.", javaHome)
		return nil
	}

	return &javaStore{
		javaHome:    javaHome,
		keytoolPath: keytoolPath,
		cacertsPath: cacertsPath,
		storePass:   "changeit", // Java's default keystore password
	}
}

func (s *javaStore) Name() string {
	return "Java Keystore"
}

func (s *javaStore) IsInstalled(cert *x509.Certificate) bool {
	// List all certificates in the keystore
	cmd := exec.Command(s.keytoolPath, "-list",
		"-keystore", s.cacertsPath,
		"-storepass", s.storePass)

	out, err := ExecCommand(cmd)
	if err != nil {
		return false
	}

	// Calculate fingerprints
	sha1Fingerprint := fmt.Sprintf("%X", sha1.Sum(cert.Raw))
	sha256Fingerprint := fmt.Sprintf("%X", sha256.Sum256(cert.Raw))

	// Remove colons from output for comparison
	outStr := strings.ReplaceAll(string(out), ":", "")

	// Check if either fingerprint exists
	return strings.Contains(outStr, sha1Fingerprint) ||
		strings.Contains(outStr, sha256Fingerprint)
}

func (s *javaStore) Install(certPath string, cert *x509.Certificate) error {
	// Import certificate to keystore
	cmd := exec.Command(s.keytoolPath, "-importcert",
		"-noprompt",
		"-keystore", s.cacertsPath,
		"-storepass", s.storePass,
		"-file", certPath,
		"-alias", CAUniqueName)

	if err := s.execKeytool(cmd); err != nil {
		return fmt.Errorf("failed to import certificate: %w", err)
	}

	return nil
}

func (s *javaStore) Uninstall(cert *x509.Certificate) error {
	// Delete certificate from keystore
	cmd := exec.Command(s.keytoolPath, "-delete",
		"-alias", CAUniqueName,
		"-keystore", s.cacertsPath,
		"-storepass", s.storePass)

	if err := s.execKeytool(cmd); err != nil {
		// Ignore errors if alias doesn't exist
		if !strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("failed to delete certificate: %w", err)
		}
	}

	return nil
}

// execKeytool executes keytool with automatic sudo retry on permission errors
func (s *javaStore) execKeytool(cmd *exec.Cmd) error {
	out, err := ExecCommand(cmd)
	if err != nil {
		// Check for permission errors
		outStr := string(out)
		if strings.Contains(outStr, "java.io.FileNotFoundException") ||
			strings.Contains(outStr, "Access is denied") {
			// Retry with sudo
			sudoCmd := CommandWithSudo(cmd.Args...)
			sudoCmd.Env = append(os.Environ(), "JAVA_HOME="+s.javaHome)
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

// findKeytool finds the keytool binary in JAVA_HOME
func findKeytool(javaHome string) string {
	keytoolName := "keytool"
	if runtime.GOOS == "windows" {
		keytoolName = "keytool.exe"
	}

	// Check bin directory
	keytoolPath := filepath.Join(javaHome, "bin", keytoolName)
	if PathExists(keytoolPath) {
		return keytoolPath
	}

	return ""
}

// findCacerts finds the cacerts keystore in JAVA_HOME
func findCacerts(javaHome string) string {
	// Try modern Java (9+) path first
	cacertsPath := filepath.Join(javaHome, "lib", "security", "cacerts")
	if PathExists(cacertsPath) {
		return cacertsPath
	}

	// Try older Java path with separate JRE
	cacertsPath = filepath.Join(javaHome, "jre", "lib", "security", "cacerts")
	if PathExists(cacertsPath) {
		return cacertsPath
	}

	return ""
}
