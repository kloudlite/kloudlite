//go:build windows

package hosts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NewWindowsManager creates a new Windows hosts manager
func NewWindowsManager() *WindowsManager {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = "C:\\Windows"
	}

	return &WindowsManager{
		BaseManager: BaseManager{
			klHostsPath:     "C:\\kloudlite\\kl.hosts",
			systemHostsPath: filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts"),
		},
	}
}

// Add adds or updates a host entry
func (m *WindowsManager) Add(hostname, ip, comment string) error {
	if err := ValidateHostname(hostname); err != nil {
		return err
	}

	if err := ValidateIP(ip); err != nil {
		return err
	}

	// Read existing entries
	entries, err := m.ReadKLHosts()
	if err != nil {
		return err
	}

	// Update or add entry
	found := false
	for i, entry := range entries {
		if entry.Hostname == hostname {
			entries[i].IP = ip
			if comment != "" {
				entries[i].Comment = comment
			}
			found = true
			break
		}
	}

	if !found {
		entries = append(entries, Entry{
			IP:       ip,
			Hostname: hostname,
			Comment:  comment,
		})
	}

	// Write kl.hosts
	if err := m.WriteKLHosts(entries); err != nil {
		return err
	}

	// Sync to system hosts
	return m.Sync()
}

// Remove removes a host entry
func (m *WindowsManager) Remove(hostname string) error {
	// Read existing entries
	entries, err := m.ReadKLHosts()
	if err != nil {
		return err
	}

	// Remove entry
	var newEntries []Entry
	found := false
	for _, entry := range entries {
		if entry.Hostname != hostname {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("hostname not found: %s", hostname)
	}

	// Write kl.hosts
	if err := m.WriteKLHosts(newEntries); err != nil {
		return err
	}

	// Sync to system hosts
	return m.Sync()
}

// List returns all kloudlite-managed entries
func (m *WindowsManager) List() ([]Entry, error) {
	return m.ReadKLHosts()
}

// Sync synchronizes kl.hosts to Windows hosts file
func (m *WindowsManager) Sync() error {
	// Backup first
	backupPath, err := m.BackupSystemHosts()
	if err != nil {
		return fmt.Errorf("failed to backup hosts file: %w", err)
	}
	fmt.Printf("Created backup: %s\n", backupPath)

	// Read kl.hosts entries
	klEntries, err := m.ReadKLHosts()
	if err != nil {
		return err
	}

	// Read system hosts
	lines, err := m.ReadSystemHosts()
	if err != nil {
		return err
	}

	// Remove existing managed section
	lines = m.RemoveManagedSection(lines)

	// Trim trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Add new managed section if there are entries
	// Windows doesn't support includes, so we append inline
	if len(klEntries) > 0 {
		lines = append(lines, "")
		lines = append(lines, BeginMarker)

		for _, entry := range klEntries {
			line := fmt.Sprintf("%-15s %s", entry.IP, entry.Hostname)
			if entry.Comment != "" {
				line += " " + entry.Comment
			}
			lines = append(lines, line)
		}

		lines = append(lines, EndMarker)
	}

	// Write system hosts
	// Windows requires proper line endings
	content := strings.Join(lines, "\r\n") + "\r\n"

	if err := os.WriteFile(m.systemHostsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write hosts file: %w (try running as Administrator)", err)
	}

	return nil
}

// Clean removes all kloudlite entries
func (m *WindowsManager) Clean() error {
	// Backup first
	backupPath, err := m.BackupSystemHosts()
	if err != nil {
		return fmt.Errorf("failed to backup hosts file: %w", err)
	}
	fmt.Printf("Created backup: %s\n", backupPath)

	// Read system hosts
	lines, err := m.ReadSystemHosts()
	if err != nil {
		return err
	}

	// Remove managed section
	lines = m.RemoveManagedSection(lines)

	// Write system hosts
	content := strings.Join(lines, "\r\n") + "\r\n"

	if err := os.WriteFile(m.systemHostsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write hosts file: %w (try running as Administrator)", err)
	}

	// Remove kl.hosts file
	if err := os.Remove(m.klHostsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove kl.hosts: %w", err)
	}

	return nil
}

// Flush flushes the DNS cache on Windows
func (m *WindowsManager) Flush() error {
	fmt.Println("Flushing DNS cache...")

	cmd := exec.Command("ipconfig", "/flushdns")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to flush DNS cache: %w\nOutput: %s", err, output)
	}

	fmt.Println("DNS cache flushed successfully")
	fmt.Println(string(output))

	return nil
}
