package wg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/ui/text"
	"github.com/kloudlite/kl/lib/util"
	"github.com/kloudlite/kl/lib/wgc"
	"github.com/spf13/cobra"
)

// not required in linux
func startServiceInBg() {
	command := exec.Command("kl", "wg", "connect", "--foreground")
	err := command.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	configFolder, err := util.GetConfigFolder()
	if err != nil {
		common_util.PrintError(err)
		return
	}

	err = os.WriteFile(configFolder+"/wgpid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
	if err != nil {
		common_util.PrintError(err)
		return
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = execCmd(fmt.Sprintf("chown %s %s", usr, configFolder+"/wgpid"),
			false); err != nil {
			common_util.PrintError(err)
			return
		}
	}
}

var foreground bool
var connectVerbose bool

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connect selected device wireguard",
	Long: `connect wireguard
Examples:
  # connect the selected device
  kl wg connect

  # connect the selected device with verbose
  kl wg connect -v

	`,
	Run: func(_ *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			common_util.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		if foreground {
			if err := startService(connectVerbose); err != nil {
				common_util.PrintError(err)
				return
			}
		}

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			common_util.PrintError(err)
			return
		}

		if strings.TrimSpace(wgInterface) != "" {
			common_util.Log("[#] already connected")

			common_util.Log("\n[#] reconnecting")

			if err := disconnect(connectVerbose); err != nil {
				common_util.PrintError(err)
				return
			}

			if err := connect(connectVerbose); err != nil {
				common_util.PrintError(err)
				return
			}

			common_util.Log("[#] connected")
			common_util.Log("[#] reconnection done")

			return
		}

		if err := connect(connectVerbose); err != nil {
			common_util.PrintError(err)
			return
		}

		// if foreground {
		// 	if err := startService(connectVerbose); err != nil {
		// 		common.PrintError(err)
		// 		return
		// 	}
		// } else {
		// 	startServiceInBg()
		// 	if err := startConfiguration(connectVerbose); err != nil {
		// 		common.PrintError(err)
		// 		return
		// 	}
		// }

		common_util.Log("[#] connected")
	},
}

func init() {
	connectCmd.Flags().BoolVar(&foreground, "foreground", false, "")
	connectCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
}
