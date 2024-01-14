package vpn

import (
	"errors"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var startFgCmd = &cobra.Command{
	Use:    "start-fg",
	Short:  "start vpn foreground",
	Hidden: true,
	Run: func(cmd *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		devName := fn.ParseStringFlag(cmd, "device")
		if devName == "" {
			fn.PrintError(errors.New("device name is required"))
			return
		}

		if err := startService(devName, false); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	startFgCmd.Flags().StringP("device", "d", "", "device name")
}
