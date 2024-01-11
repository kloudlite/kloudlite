package vpn

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "activate vpn in any environment",
	Long: `This command let you activate vpn in any environment.
Example:
  # activate vpn in any environment
  kl vpn activate
	`,
	Run: func(cmd *cobra.Command, _ []string) {

		envName := fn.ParseStringFlag(cmd, "envname")
		projectName := fn.ParseStringFlag(cmd, "projectname")

		if err := server.UpdateDeviceEnv([]fn.Option{
			fn.MakeOption("envName", envName),
			fn.MakeOption("projectName", projectName),
		}...); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("namespace updated successfully")
	},
}

func init() {
	activateCmd.Aliases = append(listCmd.Aliases, "active", "act", "a")
	activateCmd.Flags().StringP("envname", "n", "", "environment name")
	activateCmd.Flags().StringP("projectname", "p", "", "project name")
}
