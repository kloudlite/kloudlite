//go:build linux

package hosts

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
)

// NewLinuxManager creates a new Linux hosts manager
func NewLinuxManager() *LinuxManager {
	return &LinuxManager{
		BaseManager: BaseManager{
			klHostsPath:     "/etc/kl.hosts",
			systemHostsPath: "/etc/hosts",
		},
	}
}

// Add adds or updates a host entry
func (m *LinuxManager) Add(hostname, ip, comment string) error {
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
func (m *LinuxManager) Remove(hostname string) error {
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
func (m *LinuxManager) List() ([]Entry, error) {
	return m.ReadKLHosts()
}

// Sync synchronizes kl.hosts to /etc/hosts
func (m *LinuxManager) Sync() error {
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
	cmd := truststore.CommandWithSudo("tee", m.systemHostsPath)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")

	if _, err := truststore.ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	return nil
}

// Clean removes all kloudlite entries
func (m *LinuxManager) Clean() error {
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
	cmd := truststore.CommandWithSudo("tee", m.systemHostsPath)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")

	if _, err := truststore.ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	// Remove kl.hosts file
	cmd = truststore.CommandWithSudo("rm", "-f", m.klHostsPath)
	if _, err := truststore.ExecCommand(cmd); err != nil {
		return fmt.Errorf("failed to remove kl.hosts: %w", err)
	}

	return nil
}

// Flush flushes the DNS cache on Linux
func (m *LinuxManager) Flush() error {
	fmt.Println("Flushing DNS cache...")

	// Try systemd-resolved
	cmd := exec.Command("systemd-resolve", "--flush-caches")
	if err := cmd.Run(); err == nil {
		fmt.Println("DNS cache flushed successfully (systemd-resolved)")
		return nil
	}

	// Try resolvectl
	cmd = exec.Command("resolvectl", "flush-caches")
	if err := cmd.Run(); err == nil {
		fmt.Println("DNS cache flushed successfully (resolvectl)")
		return nil
	}

	// Try nscd
	cmd = truststore.CommandWithSudo("nscd", "-i", "hosts")
	if err := cmd.Run(); err == nil {
		fmt.Println("DNS cache flushed successfully (nscd)")
		return nil
	}

	fmt.Println("Note: Could not flush DNS cache automatically. You may need to:")
	fmt.Println("  - Restart systemd-resolved: sudo systemctl restart systemd-resolved")
	fmt.Println("  - Or restart nscd: sudo systemctl restart nscd")
	fmt.Println("  - Or manually: sudo systemd-resolve --flush-caches")

	return nil
}
