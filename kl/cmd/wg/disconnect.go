package wg

import (
	"errors"
	"os"
	"strings"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/wgc"
	"github.com/spf13/cobra"
)

var disconnectVerbose bool

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "disconnect selected device wireguard",
	Long: `disconnect wireguard
Examples:
  # disconnect the selected device
  kl wg disconnect

  # disconnect the selected device with verbose
  kl wg disconnect -v`,
	Run: func(_ *cobra.Command, _ []string) {

		if euid := os.Geteuid(); euid != 0 {
			common.PrintError(
				errors.New(
					color.Text("make sure you are running command with sudo", 209),
				),
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
			common.PrintError(errors.New(color.Text("[#] no device connected yet", 209)))
			return
		}

		err = stopService(disconnectVerbose)
		if err != nil {
			common.PrintError(err)
			return
		}

		common.PrintError(errors.New("[#] disconnected"))
	},
}

func init() {
	disconnectCmd.Flags().BoolVarP(&disconnectVerbose, "verbose", "v", false, "show verbose")
}
