package auth

import (
	"fmt"

	"github.com/kloudlite/kl/lib"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Long: `This command let you login to the kloudlite.
Example:
  # Login to kloudlite
  kl login 

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		err := lib.Login()
		if err != nil {
			common_util.PrintError(err)
			return
		}

		fmt.Println("successfully logged in")
	},
}
