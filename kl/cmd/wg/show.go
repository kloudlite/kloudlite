package wg

import (
	"os"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/ui/text"
	"github.com/kloudlite/kl/lib/wgc"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show connected configuration",
	Long: `show connected wireguard configuration
Examples:
  # show connected configuration
  kl wg show
	`,
	Run: func(_ *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			common_util.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		_, err := wgc.Show(nil)
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}
