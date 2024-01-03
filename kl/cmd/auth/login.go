package auth

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
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
		loginId, err := server.CreateRemoteLogin()
		if err != nil {
			functions.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fmt.Println(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(link, 21))
		fmt.Println("")

		if err = server.Login(loginId); err != nil {
			functions.PrintError(err)
			return
		}

		fmt.Println("successfully logged in")

	},
}
