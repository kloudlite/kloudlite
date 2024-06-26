package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Run: func(_ *cobra.Command, _ []string) {
		loginId, err := server.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(text.Blue(link), 21))
		fn.Log("\n")

		go func() {
			fn.Log("press enter to open link in browser")
			reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				fn.PrintError(err)
				return
			}
			if strings.Contains(reader, "\n") {
				err := fn.OpenUrl(link)
				if err != nil {
					fn.PrintError(err)
					return
				}
			} else {
				fn.Log("Invalid input\n")
			}
		}()

		if err = server.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully logged in\n")
	},
}
