package status

import (
	"fmt"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "get status of your current context (user, account, project, environment, vpn status)",
	Example: fn.Desc("{cmd} status"),
	Run: func(_ *cobra.Command, _ []string) {

		if u, err := server.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		if s, err := client.CurrentAccountName(); err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), s))
		}

		switch flags.CliName {
		case constants.CoreCliName:
			{
				if s, err := client.CurrentProjectName(); err == nil {
					fn.Log(fmt.Sprint(text.Bold(text.Blue("Project: ")), s))
				}

				if e, err := client.CurrentEnv(); err == nil {
					fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
				}
			}

		case constants.InfraCliName:
			{
				if s, err := client.CurrentClusterName(); err == nil {
					fn.Log(fmt.Sprint(text.Bold(text.Blue("Cluster: ")), s))
				}
			}
		}

		if s, err := client.CurrentDeviceName(); err == nil {

			// dev, err := server.GetDevice(fn.MakeOption("deviceName", s))
			// if err != nil {
			// 	fn.PrintError(err)
			// 	return
			// }

			// switch flags.CliName {
			// case constants.InfraCliName:
			// 	fn.Log(fmt.Sprint(text.Bold("Cluster:"), dev.ClusterName))
			// }

			b := server.CheckDeviceStatus()

			fn.Log(fmt.Sprint(text.Bold(text.Blue("Device: ")), s, func() string {
				if b {
					return text.Bold(text.Green(" (Connected) "))
				} else {
					return text.Bold(text.Red(" (Disconnected) "))
				}
			}(), text.Blue("(10.13.0.1) "), text.Blue("(10.13.0.3) ")))
			// Device IP and DNS IPv4 is same for all the devices
			// 10.13.0.1 - IP
			// 10.13.0.3 - DNS IPv4
			//fn.Log(text.Bold(text.Blue("Device IP address: ") + "10.13.0.1"))
			//fn.Log(text.Bold(text.Blue("Device DNS: ") + "10.13.0.3"))
			//resolveConfig, _ := wg_vpn.getCurrentDns(false)
			//fmt.Println(resolveConfig)
		}
	},
}
