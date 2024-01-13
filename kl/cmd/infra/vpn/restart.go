package vpn

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"os"
	"strings"
	"time"

	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var reconnectVerbose bool
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart vpn device",
	Long: `This command let you restart vpn device.
Example:
  # restart vpn device
  kl vpn restart
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
			fn.Log(text.Colored("[#] no devices connected yet", 209))
		} else {
			if err := disconnect(reconnectVerbose); err != nil {
				fn.PrintError(err)
				return
			}
			fn.Log("[#] disconnected")
		}
		fn.Log("[#] connecting")
		time.Sleep(time.Second * 1)

		// startServiceInBg()
		if err := connect(reconnectVerbose); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("[#] connected")
		fn.Log("[#] reconnection done")

		if _, err = wgc.Show(nil); err != nil {
			fn.PrintError(err)
			return
		}

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
	restartCmd.Flags().BoolVarP(&reconnectVerbose, "verbose", "v", false, "show verbose")
	restartCmd.Aliases = append(restartCmd.Aliases, "reconnect")
}
