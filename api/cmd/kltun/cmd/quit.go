package cmd

import (
	"fmt"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Disconnect from Kloudlite VPN",
	Long: `Disconnect from the active Kloudlite VPN connection.

This will stop the VPN connection, remove host entries, but keep the CA certificate
installed for future connections.`,
	Example: `  # Disconnect from VPN
  kltun quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runQuit()
	},
}

func init() {
	RootCmd.AddCommand(quitCmd)
}

func runQuit() error {
	fmt.Println("Disconnecting from Kloudlite VPN...")

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

	// Stop VPN connection via daemon
	if err := client.VPNQuit(); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Disconnected from Kloudlite VPN")
	fmt.Println()
	fmt.Println("Use 'kltun connect' to reconnect.")
	fmt.Println()

	return nil
}
