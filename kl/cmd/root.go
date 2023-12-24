package cmd

import (
	"os"

	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.CmdName,
	Short: constants.CmdName + " is command line interface to interact with kloudlite environments",
	Long: `
kl(Kloudlite Cli) will manage and attach to kloudlite environments.

Find more information at https://kloudlite.io/docs/cli

> NOTE: default kl-config file is kl.yml you can provide your own by providing KLCONFIG_PATH to the environment.`,
	// Run: runDocGen,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
