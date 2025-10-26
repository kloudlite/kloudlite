package cmd

import (
	"github.com/spf13/cobra"
)

const Version = "0.1.0"

// RootCmd represents the base command
var RootCmd = &cobra.Command{
	Use:   "kli",
	Short: "Kloudlite Installer CLI",
	Long: `kli is a command-line tool for managing Kloudlite installations.

It provides commands to create, configure, and manage Kloudlite installations
from the command line.`,
	Example: `  # Show version
  kli version
  kli v

  # Check AWS prerequisites
  kli aws doctor

  # Check GCP prerequisites
  kli gcp doctor

  # Check Azure prerequisites
  kli azure doctor
  kli az doctor`,
}

func init() {
	// Add commands
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(awsCmd)
	RootCmd.AddCommand(gcpCmd)
	RootCmd.AddCommand(azureCmd)
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
