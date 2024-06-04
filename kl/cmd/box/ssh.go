package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "get ssh access to the box",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Ssh(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	setBoxCommonFlags(sshCmd)
	sshCmd.Flags().StringP("name", "n", "", "container name")
}
