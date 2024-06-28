package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	domainutil "github.com/kloudlite/kl/domain/util"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
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

		wpath, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		err = domainutil.ConfirmBoxRestart(wpath)
		if err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func init() {
	setBoxCommonFlags(reloadCmd)
	reloadCmd.Flags().BoolP("skip-restart", "s", false, "skip restarting the box")
}
