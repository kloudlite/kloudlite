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

		// s := fn.ParseStringFlag(cmd, "name")
		s := ""
		if s == "" && len(args) > 0 {
			s = args[0]
		}

		if err := c.Info(s); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	// infoCmd.Flags().StringP("name", "n", "", "container name")
}
