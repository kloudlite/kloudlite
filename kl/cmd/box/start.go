package box

import (
	"errors"

	"github.com/kloudlite/kl/cmd/box/boxpkg"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var imageName string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start box using kl.yml configuraiton of the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		err = c.Start()
		if err != nil && errors.Is(err, boxpkg.UserCanceled) {
			fn.Log("Operation was canceled by the user")
			return
		} else if err != nil {
			fn.PrintError(err)
			return
		}
	},
}
