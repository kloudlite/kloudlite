package status

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "get status of your current context (user, account, environment, vpn status)",
	Example: fn.Desc("{cmd} status"),
	Run: func(_ *cobra.Command, _ []string) {

		if u, err := apiclient.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		if s, err := fileclient.CurrentAccountName(); err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), s))
		}

		if e, err := fileclient.CurrentEnv(); err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		}

		if s, err := fileclient.CurrentDeviceName(); err == nil {

			// dev, err := apiclient.GetDevice(fn.MakeOption("deviceName", s))
			// if err != nil {
			// 	fn.PrintError(err)
			// 	return
			// }

			// switch flags.CliName {
			// case constants.InfraCliName:
			// 	fn.Log(fmt.Sprint(text.Bold("Cluster:"), dev.ClusterName))
			// }

			b := apiclient.CheckDeviceStatus()
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Device: ")), s, func() string {
				if b {
					return text.Bold(text.Green(" (Connected) "))
				} else {
					return text.Bold(text.Red(" (Disconnected) "))
				}
			}()))

			ip, err := fileclient.CurrentDeviceIp()
			if err == nil {
				fn.Logf("%s %s\n", text.Bold(text.Blue("Device IP:")), *ip)
			}
		}
	},
}
