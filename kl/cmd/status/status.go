package status

import (
	"fmt"

	"github.com/kloudlite/kl/domain/apiclient"

	"github.com/kloudlite/kl/domain/envclient"

	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "get status of your current context (user, account, environment, vpn status)",
	Run: func(_ *cobra.Command, _ []string) {

		if u, err := apiclient.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		acc, err := fc.CurrentAccountName()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), acc))
		}

		if e, err := fileclient.CurrentEnv(); err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		}

		if envclient.InsideBox() {
			b := apiclient.CheckDeviceStatus()
			avc, err := fc.GetVpnAccountConfig(acc)
			if err != nil {
				return
			}

			fn.Log(fmt.Sprint(text.Bold(text.Blue("Device: ")), avc.DeviceName, func() string {
				if b {
					return text.Bold(text.Green(" (Connected) "))
				} else {
					return text.Bold(text.Red(" (Disconnected) "))
				}
			}()))
		}

	},
}
