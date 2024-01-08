package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

// not required in linux
func startServiceInBg() {
	command := exec.Command("kl", "wg", "start-fg")
	err := command.Start()
	if err != nil {
		fmt.Println(err)
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
  kl vpn start
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

		if strings.TrimSpace(wgInterface) != "" {
			fn.Log("[#] already connected")

			fn.Log("\n[#] reconnecting")

			if err := disconnect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			if err := connect(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}

			fn.Log("[#] connected")
			fn.Log("[#] reconnection done")

			return
		}

		startServiceInBg()

		if err := connect(connectVerbose); err != nil {
			fn.PrintError(err)
			return
		}

		// if err := startConfiguration(connectVerbose); err != nil {
		// 	fn.PrintError(err)
		// 	return
		// }

		fn.Log("[#] connected")

		wgc.Show(nil)

	},
}

func init() {
	startCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
	startCmd.Aliases = append(stopCmd.Aliases, "connect")
}
