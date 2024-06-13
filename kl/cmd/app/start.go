package app

import (
	"os"
	"runtime"

	"github.com/kloudlite/kl/app"
	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the kloudlite controller app",
	Long:  `This is internal command`,
	Run: func(cmd *cobra.Command, args []string) {

		if runtime.GOOS != constants.RuntimeWindows {
			if euid := os.Geteuid(); euid != 0 {
				fn.Log(text.Colored("make sure you are running command with sudo", 209))
				return
			}
		}

		if err := app.RunApp(cmd.Parent().Parent().Name(), cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func init() {
	startCmd.Flags().BoolP("verbose", "v", false, "run in verbose mode")
	startCmd.Flags().BoolP("foreground", "f", false, "run in foreground mode")
}
