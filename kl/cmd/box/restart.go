package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "reload",
	Short: "reload running container",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Restart(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	restartCmd.Aliases = append(restartCmd.Aliases, "restart")
	setBoxCommonFlags(restartCmd)
}
