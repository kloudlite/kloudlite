package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var uninstallCertPath string

var uninstallCACmd = &cobra.Command{
	Use:   "uninstall-ca",
	Short: "Uninstall Kloudlite CA certificate from system trust store",
	Long: `Uninstall the Kloudlite CA certificate from the system trust store.

This command delegates to the kltun daemon which runs with elevated privileges.
The daemon will be automatically started if it's not running.`,
	Example: `  # Uninstall CA certificate
  kltun uninstall-ca --cert /path/to/ca.crt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUninstallCA()
	},
}

func init() {
	uninstallCACmd.Flags().StringVar(&uninstallCertPath, "cert", "", "Path to CA certificate file (required)")
	uninstallCACmd.MarkFlagRequired("cert")
}

func runUninstallCA() error {
	// Validate certificate file exists
	if _, err := os.Stat(uninstallCertPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found: %s", uninstallCertPath)
	}

	// Get absolute path
	absPath := uninstallCertPath
	if !filepath.IsAbs(uninstallCertPath) {
		var err error
		absPath, err = filepath.Abs(uninstallCertPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	fmt.Println("Uninstalling Kloudlite CA certificate...")

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

	// Uninstall CA via daemon
	if err := client.UninstallCA(absPath); err != nil {
		return fmt.Errorf("failed to uninstall CA: %w", err)
	}

	fmt.Println("✅ CA certificate uninstalled successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Restart your browsers for the changes to take effect")
	fmt.Println()

	return nil
}
