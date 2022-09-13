package wg

import (
	"os"
	"strings"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/wgc"
	"github.com/spf13/cobra"
)

var reconnectVerbose bool

var reconnectCmd = &cobra.Command{
	Use:   "reconnect",
	Short: "reconnect the wireguard by disconnecting previous connected",
	Long: `reconnect wireguard
Examples:
  # reconnecting the selected device
  kl wg reconnect

  # reconnect the selected device with verbose
  kl wg reconnect -v
	`,
	Run: func(_ *cobra.Command, _ []string) {

		if euid := os.Geteuid(); euid != 0 {
			common.Log(
				color.Text("make sure you are running command with sudo", 209),
			)
			return
		}

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			common.PrintError(err)
			return
		}

		if strings.TrimSpace(wgInterface) == "" {
			common.Log(color.Text("[#] no devices connected yet", 209))
			common.Log("[#] connecting")
		} else {

			if err := stopService(reconnectVerbose); err != nil {
				common.PrintError(err)
				return
			}
			common.Log("[#] disconnected")

		}

		startServiceInBg()
		if err := startConfiguration(reconnectVerbose); err != nil {
			common.PrintError(err)
			return
		}

		common.Log("[#] connected")
		common.Log("[#] reconnection done")
	},
}

func init() {
	reconnectCmd.Flags().BoolVarP(&reconnectVerbose, "verbose", "v", false, "show verbose")
}
