package vpn

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
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
  sudo kl vpn status
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

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		dev, err := server.GetDevice([]fn.Option{
			fn.MakeOption("deviceName", s),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device: ")),
			text.Red(s),
		)

		if len(dev.Spec.Ports) != 0 {
			fn.Log(text.Bold(text.Green("\n[#]Exposed Ports: ")))
			for _, v := range dev.Spec.Ports {
				fn.Log(text.Blue(fmt.Sprintf("%d:%d\t", v.Port, v.TargetPort)))
			}
		} else {
			fn.Warn("No ports exposed, you can expose ports using `kl vpn expose` command")
		}

	},
}

func init() {
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
