package wg

import (
	"strings"

	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

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

		// if euid := os.Geteuid(); euid != 0 {
		// 	common_util.Log(
		// 		text.Colored("make sure you are running command with sudo", 209),
		// 	)
		// 	return
		// }

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			common_util.PrintError(err)
			return
		}

		if strings.TrimSpace(wgInterface) == "" {
			common_util.Log(text.Colored("[#] no device connected yet", 209))
			return
		}

		err = disconnect(disconnectVerbose)
		if err != nil {
			common_util.PrintError(err)
			return
		}

		common_util.Log("[#] disconnected")
	},
}

func init() {
	disconnectCmd.Flags().BoolVarP(&disconnectVerbose, "verbose", "v", false, "show verbose")
}
