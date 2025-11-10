package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var (
	hostsComment string
	hostsForce   bool
)

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Manage /etc/hosts entries for Kloudlite",
	Long: `Manage DNS entries in /etc/hosts for Kloudlite workspaces.

Entries are stored in a separate kl.hosts file and synchronized to /etc/hosts.
This command delegates to the kltun daemon which runs with elevated privileges.`,
	Example: `  # Add a host entry
  kltun hosts add workspace1.local 192.168.1.10

  # Add with comment
  kltun hosts add workspace1.local 192.168.1.10 --comment "Dev workspace"

  # Remove a host entry
  kltun hosts remove workspace1.local

  # List all entries
  kltun hosts list

  # Sync kl.hosts to /etc/hosts
  kltun hosts sync

  # Clean all Kloudlite entries
  kltun hosts clean

  # Flush DNS cache
  kltun hosts flush`,
}

var hostsAddCmd = &cobra.Command{
	Use:   "add <hostname> <ip>",
	Short: "Add or update a host entry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := args[0]
		ip := args[1]

		fmt.Printf("Adding host entry: %s → %s\n", hostname, ip)

		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// Add host entry via daemon
		if err := client.HostsAdd(hostname, ip, hostsComment); err != nil {
			return fmt.Errorf("failed to add host entry: %w", err)
		}

		fmt.Println("✓ Host entry added successfully")

		// Flush DNS cache
		if err := client.HostsFlush(); err != nil {
			fmt.Printf("Warning: failed to flush DNS cache: %v\n", err)
		}

		return nil
	},
}

var hostsRemoveCmd = &cobra.Command{
	Use:     "remove <hostname>",
	Aliases: []string{"rm", "del", "delete"},
	Short:   "Remove a host entry",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := args[0]

		fmt.Printf("Removing host entry: %s\n", hostname)

		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// Remove host entry via daemon
		if err := client.HostsRemove(hostname); err != nil {
			return fmt.Errorf("failed to remove host entry: %w", err)
		}

		fmt.Println("✓ Host entry removed successfully")

		// Flush DNS cache
		if err := client.HostsFlush(); err != nil {
			fmt.Printf("Warning: failed to flush DNS cache: %v\n", err)
		}

		return nil
	},
}

var hostsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all Kloudlite host entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// List entries via daemon
		entries, err := client.HostsList()
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No Kloudlite host entries found")
			return nil
		}

		fmt.Printf("Kloudlite Host Entries (%d total):\n\n", len(entries))

		// Use tabwriter for formatted output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "IP ADDRESS\tHOSTNAME\tCOMMENT")
		fmt.Fprintln(w, "----------\t--------\t-------")

		for _, entry := range entries {
			comment := entry.Comment
			if comment == "" {
				comment = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", entry.IP, entry.Hostname, comment)
		}

		w.Flush()

		return nil
	},
}

var hostsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize kl.hosts to /etc/hosts",
	Long: `Synchronize entries from kl.hosts to the system hosts file.

This is automatically done when adding or removing entries,
but can be run manually if needed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Synchronizing kl.hosts to system hosts file...")

		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// Sync via daemon
		if err := client.HostsSync(); err != nil {
			return fmt.Errorf("failed to sync: %w", err)
		}

		fmt.Println("✓ Synchronized successfully")

		// Flush DNS cache
		if err := client.HostsFlush(); err != nil {
			fmt.Printf("Warning: failed to flush DNS cache: %v\n", err)
		}

		return nil
	},
}

var hostsCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all Kloudlite host entries",
	Long: `Remove all Kloudlite-managed entries from /etc/hosts and delete kl.hosts file.

This will remove all workspace DNS entries managed by Kloudlite.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !hostsForce {
			fmt.Print("This will remove all Kloudlite host entries. Continue? [y/N]: ")
			var response string
			fmt.Scanln(&response)

			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		fmt.Println("Cleaning Kloudlite host entries...")

		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// Clean via daemon
		if err := client.HostsClean(); err != nil {
			return fmt.Errorf("failed to clean: %w", err)
		}

		fmt.Println("✓ All Kloudlite host entries removed")

		// Flush DNS cache
		if err := client.HostsFlush(); err != nil {
			fmt.Printf("Warning: failed to flush DNS cache: %v\n", err)
		}

		return nil
	},
}

var hostsFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Flush DNS cache",
	Long: `Flush the system DNS cache to ensure hosts file changes take effect immediately.

Platform-specific commands:
  - macOS: dscacheutil -flushcache && killall -HUP mDNSResponder
  - Linux: systemd-resolve --flush-caches or nscd -i hosts
  - Windows: ipconfig /flushdns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure daemon is running
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.EnsureRunning(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Connect to daemon
		client := daemon.NewClient(sm.GetSocketPath())

		// Flush via daemon
		if err := client.HostsFlush(); err != nil {
			return fmt.Errorf("failed to flush DNS cache: %w", err)
		}

		return nil
	},
}

func init() {
	// Add flags
	hostsAddCmd.Flags().StringVar(&hostsComment, "comment", "", "Optional comment for the entry")
	hostsCleanCmd.Flags().BoolVarP(&hostsForce, "force", "f", false, "Skip confirmation prompt")

	// Add subcommands
	hostsCmd.AddCommand(hostsAddCmd)
	hostsCmd.AddCommand(hostsRemoveCmd)
	hostsCmd.AddCommand(hostsListCmd)
	hostsCmd.AddCommand(hostsSyncCmd)
	hostsCmd.AddCommand(hostsCleanCmd)
	hostsCmd.AddCommand(hostsFlushCmd)

	// Register with root
	RootCmd.AddCommand(hostsCmd)
}
