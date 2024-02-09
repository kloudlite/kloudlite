package app

import (
	"os"

	"github.com/kloudlite/kl/app"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden: true,
	Use:    "start-app",
	Short:  "start the kloudlite app",
	Long:   `This is internal command`,
	Run: func(_ *cobra.Command, _ []string) {

		if euid := os.Geteuid(); euid != 0 {
			fn.Log(text.Colored("make sure you are running command with sudo", 209))
			return
		}

		if err := app.RunApp(); err != nil {
			fn.PrintError(err)
		}
	},
}
