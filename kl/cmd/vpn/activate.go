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
		ns := fn.ParseStringFlag(cmd, "namespace")

		if ns == "" {
			fn.Log("namespace is missing, please provide using kl infra vpn activate -n <namespace>")
			return
		}
		if err := server.UpdateDeviceNS(ns); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("namespace updated successfully")
	},
}

func init() {
	activateCmd.Aliases = append(activateCmd.Aliases, "active", "act", "a")
	activateCmd.Flags().StringP("namespace", "n", "", "namespace")
}
