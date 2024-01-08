package intercept

import (
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long: `intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl vpn intercept --app <app_name>

	# stop intercept app with selected vpn device
	kl vpn intercept stop -app <app_name>

  # intercept app with specified vpn device
  kl vpn intercept --app <app_name> --device <device_name>
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
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.Aliases = append(startCmd.Aliases, "inc")
}
