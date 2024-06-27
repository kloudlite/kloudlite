package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/fileclient"

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

		klfile, err := fileclient.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Start(klfile); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	setBoxCommonFlags(startCmd)
}
