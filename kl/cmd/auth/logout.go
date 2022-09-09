package auth

import (
	"fmt"

	"github.com/kloudlite/kl/lib"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Long: `This command will logout your session from the cli
Example:
  # Logout from kloudlite
  kl auth logout

  using above command you can expire your session from the current device shell.

  logging out will not delete your mounted config/secret files or kl-config file.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		err := lib.Logout()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(`successfully logged out.

but the mounted configs, secrets and kl-config stil there, so if you want to also clear this you have clean these manually. 
		`)
	},
}
