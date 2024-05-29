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
	Run: func(c *cobra.Command, _ []string) {

		if runtime.GOOS != constants.RuntimeWindows {
			if euid := os.Geteuid(); euid != 0 {
				fn.Log(text.Colored("make sure you are running command with sudo", 209))
				return
			}
		}

		if err := app.RunApp(c.Parent().Name()); err != nil {
			fn.PrintError(err)
		}
	},
}
