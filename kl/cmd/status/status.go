package status

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "get status of your current context (user, account, project, environment, vpn status)",
	Example: functions.Desc("{cmd} status"),
	Run: func(_ *cobra.Command, _ []string) {

		if s, err := client.CurrentAccountName(); err == nil {
			functions.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), s))
		}

		switch flags.CliName {
		case constants.CoreCliName:
			{
				if s, err := client.CurrentProjectName(); err == nil {
					functions.Log(fmt.Sprint(text.Bold(text.Blue("Project: ")), s))
				}

				if e, err := client.CurrentEnv(); err == nil {
					functions.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
				}
			}

		case constants.InfraCliName:
			{
				if s, err := client.CurrentClusterName(); err == nil {
					functions.Log(fmt.Sprint(text.Bold(text.Blue("Cluster: ")), s))
				}
			}
		}

		if s, err := client.CurrentDeviceName(); err == nil {

			b := server.CheckDeviceStatus()

			functions.Log(fmt.Sprint(text.Bold(text.Blue("Device: ")), s, func() string {
				if b {
					return text.Bold(text.Green(" (Connected)"))
				} else {
					return text.Bold(text.Red(" (Disconnected)"))
				}
			}()))
		}
	},
}
