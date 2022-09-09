package auth

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login | logout to kloudlite",
	Long: `login/logout to kloudlite.

  Examples:
  # login to kloudlite
  kl auth login

  # logout to kloudlite
  kl auth logout`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
}
