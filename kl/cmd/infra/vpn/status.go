package vpn

import (
	"os"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show vpn status",
	Long: `This command let you show vpn status.
Example:
  # show vpn status
  sudo kl infra vpn status
	`,
	Run: func(_ *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		_, err := wgc.Show(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentInfraDeviceName()
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
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
