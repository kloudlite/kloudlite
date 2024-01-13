package intercept

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop intercept app to stop tunnel trafic to your device",
	Long: `stop intercept app to stop tunnel trafic to your device
Examples:
	# close intercept app
  kl vpn intercept stop --app <app_name>
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		ns := ""

		if cmd.Flags().Changed("name") {
			ns, _ = cmd.Flags().GetString("name")
		}

		if ns == "" {
			e, err := server.EnsureEnv(nil)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ns = e.TargetNs
		}

		if err := server.UpdateDeviceNS(ns); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("namespace updated successfully")
	},
}

func init() {
	stopCmd.Flags().StringP("app", "a", "", "app name")

	stopCmd.Aliases = append(startCmd.Aliases, "close", "end", "leave", "quit", "terminate", "exit")
}
