package wg

import (
	"errors"
	"os"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
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
			common.PrintError(
				errors.New(
					color.ColorText("make sure you are running command with sudo", 209),
				),
			)
			return
		}

		if err := stopService(reconnectVerbose); err != nil {
			common.PrintError(err)
			return
		}
		common.PrintError(errors.New("[#] disconnected"))

		startServiceInBg()
		if err := startConfiguration(reconnectVerbose); err != nil {
			common.PrintError(err)
			return
		}

		common.PrintError(errors.New("[#] connected"))
		common.PrintError(errors.New("[#] reconnection done"))
	},
}

func init() {
	reconnectCmd.Flags().BoolVarP(&reconnectVerbose, "verbose", "v", false, "show verbose")
}
