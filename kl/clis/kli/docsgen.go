package kli

import (
	"github.com/kloudlite/kl/clis"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var DocsCmd = &cobra.Command{
	Hidden: true,
	Use:    "docs",
	Short:  "generate docs for kloudlite cli",
	Long: `This command let you generate docs for kloudlite cli.

Example:
  # generate docs for kloudlite cli
  kl docs

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(_ *cobra.Command, args []string) {
		if err := clis.RunDocGen(rootCmd, args); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully generated docs/kli")
	},
}
