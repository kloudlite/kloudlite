package cmd

import (
	"errors"
	"os"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib/common"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.CMD_NAME,
	Short: constants.CMD_NAME + " is command line interface to interact with kloudlite environments",
	Long: `
kl(Kloudlite Cli) will manage and attach to kloudlite environments.

Find more information at https://kloudlite.io/docs/cli

NOTE: default kl-config file is kl.yml you can provide your own by providing KLCONFIG_PATH to the environment.
	`,
	// Run: run,
}

func run(cmd *cobra.Command, _ []string) {

	if _, er := os.Stat("./docs"); errors.Is(er, os.ErrNotExist) {
		err := os.MkdirAll("./docs", os.ModePerm)
		if err != nil {
			common.PrintError(err)
			return
		}

	} else {
		err := os.RemoveAll("./docs")

		if err != nil {
			common.PrintError(err)
			return
		}

		err = os.MkdirAll("./docs", os.ModePerm)
		if er != nil {
			common.PrintError(err)
			return
		}
	}

	err := doc.GenMarkdownTree(cmd, "./docs")
	if err != nil {
		common.PrintError(err)
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
