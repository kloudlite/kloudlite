package wg

import (
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"os"
	"strings"
	"time"

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
			common_util.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			common_util.PrintError(err)
			return
		}

		if strings.TrimSpace(wgInterface) == "" {
			common_util.Log(text.Colored("[#] no devices connected yet", 209))
		} else {
			if err := disconnect(reconnectVerbose); err != nil {
				common_util.PrintError(err)
				return
			}
			common_util.Log("[#] disconnected")
		}
		common_util.Log("[#] connecting")
		time.Sleep(time.Second * 1)

		// startServiceInBg()
		if err := connect(reconnectVerbose); err != nil {
			common_util.PrintError(err)
			return
		}

		common_util.Log("[#] connected")
		common_util.Log("[#] reconnection done")
	},
}

func init() {
	reconnectCmd.Flags().BoolVarP(&reconnectVerbose, "verbose", "v", false, "show verbose")
}
