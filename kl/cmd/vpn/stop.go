package vpn

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"os"
	"strings"

	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var disconnectVerbose bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop vpn device",
	Long: `This command let you stop running vpn device.
Example:
  # stop vpn device
  sudo kl vpn stop
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

		if strings.TrimSpace(wgInterface) == "" {
			fn.Log(text.Colored("[#] no device connected yet", 209))
			return
		}

		err = disconnect(disconnectVerbose)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("[#] disconnected")

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\n[#]Selected Device:")),
			text.Red(s),
		)
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&disconnectVerbose, "verbose", "v", false, "show verbose")

	stopCmd.Aliases = append(stopCmd.Aliases, "disconnect")
}
