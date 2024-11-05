package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "get info about a container",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Info(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}
