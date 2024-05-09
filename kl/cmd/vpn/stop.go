package vpn

import (
	"os"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop vpn device",
	Long: `This command let you stop running vpn device.
Example:
  # stop vpn device
  sudo kl vpn stop
	`,
	Run: func(cmd *cobra.Command, _ []string) {

		verbose := fn.ParseBoolFlag(cmd, "verbose")

		if runtime.GOOS == constants.RuntimeWindows {
			if err := disconnect(verbose); err != nil {
				fn.Notify("Error:", err.Error())
				fn.PrintError(err)
			}
			return
		}

		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(wgInterface) == 0 {
			fn.Log(text.Colored("[#] no device connected yet", 209))
			return
		}

		err = disconnect(verbose)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("[#] disconnected")

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device: ")),
			text.Red(s),
		)
	},
}

func init() {
	stopCmd.Flags().BoolP("verbose", "v", false, "run in debug mode")

	stopCmd.Aliases = append(stopCmd.Aliases, "disconnect")
}
