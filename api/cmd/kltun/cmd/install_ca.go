package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var installCertPath string

var installCACmd = &cobra.Command{
	Use:   "install-ca",
	Short: "Install Kloudlite CA certificate to system trust store",
	Long: `Install the Kloudlite CA certificate to the system trust store.

This command delegates to the kltun daemon which runs with elevated privileges.
The daemon will be automatically started if it's not running.`,
	Example: `  # Install CA certificate
  kltun install-ca --cert /path/to/ca.crt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstallCA()
	},
}

func init() {
	installCACmd.Flags().StringVar(&installCertPath, "cert", "", "Path to CA certificate file (required)")
	installCACmd.MarkFlagRequired("cert")
}

func runInstallCA() error {
	// Validate certificate file exists
	if _, err := os.Stat(installCertPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found: %s", installCertPath)
	}

	// Get absolute path
	absPath := installCertPath
	if !filepath.IsAbs(installCertPath) {
		var err error
		absPath, err = filepath.Abs(installCertPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	fmt.Println("Installing Kloudlite CA certificate...")

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

	// Install CA via daemon
	if err := client.InstallCA(absPath); err != nil {
		return fmt.Errorf("failed to install CA: %w", err)
	}

	fmt.Println("✅ CA certificate installed successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Restart your browsers for the changes to take effect")
	fmt.Println("  - Test with: curl https://your-kloudlite-domain.com")
	fmt.Println()

	return nil
}
