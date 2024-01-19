package vpn

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

// not required in linux
func startServiceInBg(devName string) {
	command := exec.Command("kl", "vpn", "start-fg", "-d", devName)
	err := command.Start()
	if err != nil {
		fn.Log(err)
		return
	}
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		fn.PrintError(err)
		return
	}

	err = os.WriteFile(configFolder+"/wgpid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
	if err != nil {
		fn.PrintError(err)
		return
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = execCmd(fmt.Sprintf("chown %s %s", usr, configFolder+"/wgpid"),
			false); err != nil {
			fn.PrintError(err)
			return
		}
	}
}

var foreground bool
var connectVerbose bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn device",
	Long: `This command let you start vpn device.
Example:
  # start vpn device
  sudo kl infra vpn start
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

		if len(wgInterface) != 0 {
			fn.Log("[#] already connected")

			fn.Log("\n[#] reconnecting")

			if err := disconnect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			devName, err := client.CurrentInfraDeviceName()
			if err != nil {
				fn.PrintError(err)
				return
			}

			startServiceInBg(devName)

			if err := connect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			fn.Log("[#] connected")
			fn.Log("[#] reconnection done")

			return
		}

		devName, err := client.CurrentInfraDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		startServiceInBg(devName)

		if err := connect(connectVerbose); err != nil {
			fn.PrintError(err)
			return
		}

		// if err := startConfiguration(connectVerbose); err != nil {
		// 	fn.PrintError(err)
		// 	return
		// }

		fn.Log("[#] connected")

		_, err = wgc.Show(nil)

		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentInfraDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device:")),
			text.Red(s),
		)
	},
}

func init() {
	startCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
	startCmd.Aliases = append(stopCmd.Aliases, "connect")
}
