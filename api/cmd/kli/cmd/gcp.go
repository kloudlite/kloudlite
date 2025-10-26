package cmd

import (
	"github.com/spf13/cobra"
)

// gcpCmd represents the gcp command
var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "GCP provider commands",
	Long: `Manage Kloudlite installations on Google Cloud Platform.

This command provides subcommands for installing, configuring, and managing
Kloudlite on GCP.`,
	Example: `  # Check GCP prerequisites
  kli gcp doctor

  # Install Kloudlite on GCP
  kli gcp install`,
}

func init() {
	// Add GCP subcommands
	gcpCmd.AddCommand(gcpDoctorCmd)
}
