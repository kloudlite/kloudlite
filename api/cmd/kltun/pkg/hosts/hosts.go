package hosts

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Markers for managed section in /etc/hosts
	BeginMarker = "# BEGIN KLOUDLITE MANAGED HOSTS - DO NOT EDIT"
	EndMarker   = "# END KLOUDLITE MANAGED HOSTS"

	// Comment header for kl.hosts file
	KLHostsHeader = `# Kloudlite managed hosts - DO NOT EDIT MANUALLY
# Managed by kltun - changes will be overwritten
# Last updated: %s
`
)

// Entry represents a host entry
type Entry struct {
	IP       string
	Hostname string
	Comment  string
}

// Manager manages hosts file entries
type Manager interface {
	// Add adds or updates a host entry
	Add(hostname, ip, comment string) error

	// Remove removes a host entry
	Remove(hostname string) error

	// List returns all kloudlite-managed entries
	List() ([]Entry, error)

	// Sync synchronizes kl.hosts to /etc/hosts
	Sync() error

	// Clean removes all kloudlite entries
	Clean() error

	// Flush flushes the DNS cache
	Flush() error

	// GetKLHostsPath returns the path to kl.hosts file
	GetKLHostsPath() string

	// GetSystemHostsPath returns the path to system hosts file
	GetSystemHostsPath() string
}

// BaseManager provides common functionality
type BaseManager struct {
	klHostsPath     string
	systemHostsPath string
}

// ValidateHostname validates a hostname
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	// Simple validation - alphanumeric, dots, hyphens
	for _, char := range hostname {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-') {
			return fmt.Errorf("invalid character in hostname: %c", char)
		}
	}

	return nil
}

// ValidateIP validates an IP address
func ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// ReadKLHosts reads entries from kl.hosts file
func (bm *BaseManager) ReadKLHosts() ([]Entry, error) {
	file, err := os.Open(bm.klHostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("failed to open kl.hosts: %w", err)
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse entry
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		entry := Entry{
			IP:       parts[0],
			Hostname: parts[1],
		}

		// Check for inline comment
		if len(parts) > 2 && strings.HasPrefix(parts[2], "#") {
			entry.Comment = strings.Join(parts[2:], " ")
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading kl.hosts: %w", err)
	}

	return entries, nil
}

// WriteKLHosts writes entries to kl.hosts file
func (bm *BaseManager) WriteKLHosts(entries []Entry) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(bm.klHostsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first
	tmpFile := bm.klHostsPath + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write header
	header := fmt.Sprintf(KLHostsHeader, time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(header + "\n"); err != nil {
		file.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		line := fmt.Sprintf("%-15s %s", entry.IP, entry.Hostname)
		if entry.Comment != "" {
			line += " " + entry.Comment
		}
		line += "\n"

		if _, err := file.WriteString(line); err != nil {
			file.Close()
			os.Remove(tmpFile)
			return fmt.Errorf("failed to write entry: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, bm.klHostsPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// BackupSystemHosts creates a backup of the system hosts file
func (bm *BaseManager) BackupSystemHosts() (string, error) {
	backupPath := bm.systemHostsPath + ".klbackup." + time.Now().Format("20060102-150405")

	src, err := os.Open(bm.systemHostsPath)
	if err != nil {
		return "", fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return backupPath, nil
}

// ReadSystemHosts reads the system hosts file
func (bm *BaseManager) ReadSystemHosts() ([]string, error) {
	file, err := os.Open(bm.systemHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}

	return lines, nil
}

// WriteSystemHosts writes lines to the system hosts file
func (bm *BaseManager) WriteSystemHosts(lines []string) error {
	// Write to temporary file first
	tmpFile := bm.systemHostsPath + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			file.Close()
			os.Remove(tmpFile)
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, bm.systemHostsPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// RemoveManagedSection removes the managed section from system hosts file
func (bm *BaseManager) RemoveManagedSection(lines []string) []string {
	var result []string
	inManagedSection := false

	for _, line := range lines {
		if strings.Contains(line, BeginMarker) {
			inManagedSection = true
			continue
		}

		if strings.Contains(line, EndMarker) {
			inManagedSection = false
			continue
		}

		if !inManagedSection {
			result = append(result, line)
		}
	}

	return result
}

// GetKLHostsPath returns the path to kl.hosts
func (bm *BaseManager) GetKLHostsPath() string {
	return bm.klHostsPath
}

// GetSystemHostsPath returns the path to system hosts file
func (bm *BaseManager) GetSystemHostsPath() string {
	return bm.systemHostsPath
}

// Platform-specific manager types (defined in platform-specific files)
type DarwinManager struct {
	BaseManager
}

type LinuxManager struct {
	BaseManager
}

type WindowsManager struct {
	BaseManager
}

// NewManager is implemented in platform-specific manager_*.go files
