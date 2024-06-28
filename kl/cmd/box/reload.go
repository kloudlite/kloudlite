package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "reload the box according to the current kl.yml configuration",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Reload(); err != nil {
			fn.PrintError(err)
			return
		}

		if err = c.ConfirmBoxRestart(); err != nil {
			fn.PrintError(err)
			return
		}

	},
}
