package app

import (
	"github.com/kloudlite/kl/app"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden: true,
	Use:    "start-app",
	Short:  "start the kloudlite app",
	Long:   `This is internal command`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := app.RunApp(); err != nil {
			functions.PrintError(err)
		}
	},
}
