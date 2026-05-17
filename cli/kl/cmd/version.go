package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kl version %s\n", Version)
	},
}
