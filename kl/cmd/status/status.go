package status

import (
	"errors"
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
	Run: func(cmd *cobra.Command, _ []string) {

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if u, err := apic.GetCurrentUser(); err == nil {
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

		e, err := fc.CurrentEnv()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		} else if errors.Is(err, fileclient.NoEnvSelected) {
			filePath := fn.ParseKlFile(cmd)
			klFile, err := fc.GetKlFile(filePath)
			if err != nil {
				fn.PrintError(err)
				return
			}
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), klFile.DefaultEnv))
		}

		if envclient.InsideBox() {
			b := apic.CheckDeviceStatus()
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
