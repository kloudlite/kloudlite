package auth

import (
	"github.com/kloudlite/kl/lib"
	common_util "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var WhoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "session check for kloudlite",
	Long: `This command let you login to the kloudlite.
Example:
  # Login to kloudlite
  kl whoami 

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		err := lib.WhoAmI()
		if err != nil {
			common_util.PrintError(err)
			return
		}
	},
}
