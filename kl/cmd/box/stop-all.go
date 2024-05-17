package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopAllCmd = &cobra.Command{
	Hidden: true,
	Use:    "stop-all",
	Short:  "stop all running boxes",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.StopAll(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	setBoxCommonFlags(stopAllCmd)
}
