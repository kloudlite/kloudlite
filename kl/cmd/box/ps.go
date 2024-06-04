package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all running boxes",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.ListBox(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	psCmd.Aliases = append(psCmd.Aliases, "ls")
	setBoxCommonFlags(psCmd)
}
