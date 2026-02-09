package cmd

import (
	"github.com/spf13/cobra"
)

// ociCmd represents the oci command
var ociCmd = &cobra.Command{
	Use:   "oci",
	Short: "OCI provider commands",
	Long: `Manage Kloudlite installations on Oracle Cloud Infrastructure.

This command provides subcommands for installing, configuring, and managing
Kloudlite on OCI.`,
	Example: `  # Check OCI prerequisites
  kli oci doctor

  # Install Kloudlite on OCI
  kli oci install`,
}

func init() {
	// Add OCI subcommands
	ociCmd.AddCommand(ociDoctorCmd)
	ociCmd.AddCommand(ociInstallCmd)
	ociCmd.AddCommand(ociUninstallCmd)
}
