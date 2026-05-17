package cmd

import (
	"github.com/spf13/cobra"
)

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS provider commands",
	Long: `Manage Kloudlite installations on AWS.

This command provides subcommands for installing, configuring, and managing
Kloudlite on Amazon Web Services.`,
	Example: `  # Check AWS prerequisites
  kli aws doctor

  # Install Kloudlite on AWS
  kli aws install`,
}

func init() {
	// Add AWS subcommands
	awsCmd.AddCommand(awsDoctorCmd)
	awsCmd.AddCommand(awsInstallCmd)
	awsCmd.AddCommand(awsUninstallCmd)
}
