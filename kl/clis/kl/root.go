package kl

import (
	"os"

	domain_util "github.com/kloudlite/kl/domain/util"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                flags.CliName,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		if (len(args) != 0) && (args[0] == "--version" || args[0] == "-v") {
			fn.Log(cmd.Version)
			return
		}

		if len(args) < 2 || args[0] != "--" {
			// fn.Log(GetRootHelp(cmd))

			if err := cmd.Help(); err != nil {
				fn.Log(err)
				os.Exit(1)
			}
			return
		}

		if err := domain_util.MountEnv(args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = flags.Version
}
