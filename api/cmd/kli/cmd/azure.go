package cmd

import (
	"github.com/spf13/cobra"
)

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
	Use:     "azure",
	Aliases: []string{"az"},
	Short:   "Azure provider commands",
	Long: `Manage Kloudlite installations on Microsoft Azure.

This command provides subcommands for installing, configuring, and managing
Kloudlite on Azure.`,
	Example: `  # Check Azure prerequisites
  kli azure doctor
  kli az doctor

  # Install Kloudlite on Azure
  kli azure install
  kli az install`,
}

func init() {
	// Add Azure subcommands
	azureCmd.AddCommand(azureDoctorCmd)
}
