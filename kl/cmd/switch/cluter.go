package sw

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:     "cluster",
	Short:   "Switch cluster",
	Example: `kl switch cluster`,
	Run: func(cmd *cobra.Command, _ []string) {
		accountName := fn.ParseStringFlag(cmd, "account")
		clusterName := fn.ParseStringFlag(cmd, "cluster")
		deviceRunning := server.CheckDeviceStatus()

		if accountName != "" {
			acc, err := server.SelectAccount(accountName)
			if err != nil {
				fn.PrintError(err)
				return
			}

			mc, err := client.GetMainCtx()
			if err != nil {
				fn.PrintError(err)
				return
			}

			if mc.AccountName != acc.Metadata.Name {
				if err := client.SetAccountToMainCtx(acc.Metadata.Name); err != nil {
					fn.PrintError(err)
					return
				}
			}
		}

		c, err := server.SelectCluster(clusterName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := client.SetClusterToMainCtx(c.Metadata.Name); err != nil {
			fn.PrintError(err)
			return
		}

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := server.UpdateDeviceClusterName(c.Metadata.Name); err != nil {
			fn.PrintError(err)
			return
		}

		if deviceRunning {
			fn.Log(text.Yellow("[#] vpn also switched to diffrent cluster, please restart vpn manually"))
		}

	},
}

func init() {
	clusterCmd.Flags().StringP("account", "a", "", "account name")
	clusterCmd.Flags().StringP("cluster", "c", "", "cluster name")
	clusterCmd.Aliases = append(accCmd.Aliases, "clus")
}
