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

		err = c.Stop()
		if err != nil {
			fn.PrintError(err)
			return
		}

		err = c.Start()
		if err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func init() {
	setBoxCommonFlags(restartCmd)
}
