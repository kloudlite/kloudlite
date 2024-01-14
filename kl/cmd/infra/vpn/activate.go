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
  kl infra vpn activate -n <namespace>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		ns := ""

		if cmd.Flags().Changed("name") {
			ns, _ = cmd.Flags().GetString("name")
		}

		if err := server.UpdateInfraDeviceNS(ns); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("namespace updated successfully")
	},
}

func init() {
	activateCmd.Aliases = append(listCmd.Aliases, "active", "act", "a")
	activateCmd.Flags().StringP("name", "n", "", "namespace")
}
