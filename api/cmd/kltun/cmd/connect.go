package cmd

import (
	"fmt"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var (
	connectToken  string
	connectServer string
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Kloudlite VPN",
	Long: `Connect to Kloudlite VPN using your authentication token.

The connection runs in the background via the kltun daemon. Once connected,
you can access your Kloudlite workspace and services.

IMPORTANT: For security reasons, credentials are NOT saved to disk. You must
provide --token and --server flags on every connection.`,
	Example: `  # Connect with token and server (required every time)
  kltun connect --token YOUR_TOKEN --server https://subdomain.khost.dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConnect()
	},
}

func init() {
	connectCmd.Flags().StringVar(&connectToken, "token", "", "Authentication token")
	connectCmd.Flags().StringVar(&connectServer, "server", "", "Server URL (e.g., https://subdomain.khost.dev)")

	RootCmd.AddCommand(connectCmd)
}

func runConnect() error {
	fmt.Println("Connecting to Kloudlite VPN...")

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

	// Start VPN connection via daemon
	sessionID, err := client.VPNConnect(connectToken, connectServer)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Connected to Kloudlite VPN!")
	fmt.Printf("  Session ID: %s\n", sessionID)
	fmt.Println()
	fmt.Println("Your VPN connection is now running in the background.")
	fmt.Println("Use 'kltun quit' to disconnect.")
	fmt.Println()

	return nil
}
