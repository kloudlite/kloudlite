package auth

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Example: fn.Desc(`# Login to kloudlite
{cmd} auth login`),
	Run: func(_ *cobra.Command, _ []string) {
		loginId, err := server.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := text.Blue(fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId))

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(link, 21))
		fn.Log("\n")

		if err = server.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully logged in\n")

	},
}
