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
	Long: `Kloudlite CLI
This cli will help you to perform all the mojor tasks. some of the tasks listed below:

  - It will help you to get configs/secrets
  - It allows you to intecept apps/lambdas with locally running app. 
  - It helps to get list of all accounts, projects, configs, secrets, devices, apps/lambdas. 
  - It helps you to manage kl-config

NOTE: default kl-config file is kl.yml you can provide your own by providing KLCONFIG_PATH to the environment.
	`,
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
