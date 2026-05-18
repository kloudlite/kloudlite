package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

// RootCmd represents the base command
var RootCmd = &cobra.Command{
	Use:   "kltun",
	Short: "Kloudlite Tunnel Manager",
	Long: `kltun is a CLI tool for managing secure tunnels to Kloudlite workspaces and services.

It provides commands to set up secure connections, install CA certificates,
and manage tunnel configurations.`,
	Example: `  # Connect to Kloudlite workspace
  kltun connect --server https://workspace.kloudlite.io

  # Install CA certificate to system trust store
  kltun install-ca --cert /path/to/ca.crt

  # Uninstall CA certificate
  kltun uninstall-ca

  # Show version
  kltun version`,
}

func init() {
	// Add commands
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(installCACmd)
	RootCmd.AddCommand(uninstallCACmd)
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kltun version %s\n", Version)
	},
}
