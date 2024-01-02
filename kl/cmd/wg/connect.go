package wg

import (
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"os"
	"os/exec"
	"strings"

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
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}
		/*
		   steps to perform
		   1. setup interface
		   2. setup dns

		*/

		if foreground {
			if err := startService(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}
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

		if err := connect(connectVerbose); err != nil {
			fn.PrintError(err)
			return
		}

		if foreground {
			if err := startService(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}
		} else {
			startServiceInBg()
			if err := startConfiguration(connectVerbose); err != nil {
				fn.PrintError(err)
				return
			}
		}

		fn.Log("[#] connected")
	},
}

func init() {
	connectCmd.Flags().BoolVar(&foreground, "foreground", false, "")
	connectCmd.Flags().BoolVarP(&connectVerbose, "verbose", "v", false, "show verbose")
}
