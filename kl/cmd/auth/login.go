package auth

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Long: `This command let you login to the kloudlite.
Example:
  # Login to kloudlite
  kl auth login 

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		loginId, err := server.CreateRemoteLogin()
		if err != nil {
			functions.PrintError(err)
			return
		}

		link := text.Blue(fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId))

		functions.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(link, 21))
		functions.Log("\n")

		if err = server.Login(loginId); err != nil {
			functions.PrintError(err)
			return
		}

		functions.Log("successfully logged in\n")

	},
}
