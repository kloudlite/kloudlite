package auth

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login and logout",
	Long: `Login and logout
Examples:
  # login to kloudlite
  kl auth login

  # logout from kloudlite
  kl auth logout
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(authStatusCmd)
}
