package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart the box according to the current kl.yml configuration",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err = c.Stop(); err != nil {
			fn.PrintError(err)
			return
		}

		if err = c.Start(); err != nil {
			fn.PrintError(err)
			return
		}

	},
}
