package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Display the CLI version",
	Long:    `Display the current version of the kli CLI tool.`,
	Example: `  kli version
  kli v`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kli version %s\n", Version)
	},
}
