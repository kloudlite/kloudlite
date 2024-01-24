package auth

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login,logout and status",
	Long: `Login, logout and status
Examples:
  # login to kloudlite
  kl auth login

  # logout from kloudlite
  kl auth logout

  # get auth status
  kl auth status
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(authStatusCmd)
}
