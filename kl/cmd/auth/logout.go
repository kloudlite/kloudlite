package auth

import (
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/domain/client"
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
		err := client.Logout()
		if err != nil {
			fn.Log(err)
			return
		}
		fn.Log(`successfully logged out.

but the mounted configs, secrets and kl-config stil there, so if you want to also clear this you have clean these manually. 
		`)
	},
}
