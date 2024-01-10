package vpn

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var startFgCmd = &cobra.Command{
	Use:    "start-fg",
	Short:  "start vpn foreground",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		if euid := os.Geteuid(); euid != 0 {
			fn.Log(
				text.Colored("make sure you are running command with sudo", 209),
			)
			return
		}

		if err := startService(false); err != nil {
			fn.PrintError(err)
			return
		}
	},
}
