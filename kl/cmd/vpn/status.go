package vpn

import (
	"os"
	"runtime"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Hidden: true,
	Use:    "status",
	Short:  "show vpn status",
	Long: `This command let you show vpn status.
Example:
  # show vpn status
  sudo kl vpn status
	`,
	Run: func(cmd *cobra.Command, _ []string) {

		if runtime.GOOS != "windows" {
			if euid := os.Geteuid(); euid != 0 {
				fn.Log(
					text.Colored("make sure you are running command with sudo", 209),
				)
				return
			}
		}

		_, err := wgc.Show(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err == nil {
			fn.Log(text.Bold(text.Green("\n[#]Selected Device: ")), text.Red(s), "\n")
		}
	},
}

func init() {
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
