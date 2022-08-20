package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/constants"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.CMD_NAME,
	Short: constants.CMD_NAME + " is command line interface to interact with kloudlite environments",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// func init() {
// 	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
