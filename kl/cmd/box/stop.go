package box

import (
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop running box",
	Run: func(cmd *cobra.Command, args []string) {

		fn.Logf(text.Yellow("[#] this action will stop the current workspace. this will end all current running processes in the container. do you want to do you want to proceed? [Y/n] "))
		if !fn.Confirm(strings.ToUpper("Y"), strings.ToUpper("Y")) {
			return
		}

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Stop(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}
