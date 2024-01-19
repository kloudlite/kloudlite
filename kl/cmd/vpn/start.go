package vpn

import (
	"os"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

// not required in linux

var foreground bool
var connectVerbose bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn device",
	Long: `This command let you start vpn device.
Example:
  # start vpn device
  sudo kl vpn start
	`,
	Run: func(_ *cobra.Command, _ []string) {
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

		if len(wgInterface) != 0 {
			fn.Log("[#] already connected")

			fn.Log("\n[#] reconnecting")

			if err := disconnect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			if err := startConnecting(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			fn.Log("[#] connected")
			fn.Log("[#] reconnection done")

			return
		}

		if err := startConnecting(connectVerbose); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("[#] connected")

		_, err = wgc.Show(nil)

		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device:")),
			text.Red(s),
		)
	},
}

func startConnecting(verbose bool) error {
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	if err := wg_vpn.StartServiceInBg(devName, configFolder); err != nil {
		return err
	}

	if err := connect(verbose); err != nil {
		return err
	}

	return nil
}

func init() {
	startCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
	startCmd.Aliases = append(stopCmd.Aliases, "connect")
}
