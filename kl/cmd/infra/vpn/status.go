package vpn

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/lib/wgc"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show vpn status",
	Long: `This command let you show vpn status.
Example:
  # show vpn status
  sudo kl vpn status
	`,
	Run: func(_ *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			functions.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		_, err := wgc.Show(nil)
		if err != nil {
			functions.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			functions.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\n[#]Selected Device:")),
			text.Red(s),
		)

	},
}

func init() {
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
