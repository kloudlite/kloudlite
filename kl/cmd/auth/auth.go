package auth

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login,logout and status",
	Example: fn.Desc(`# login to kloudlite
{cmd} auth login

# logout from kloudlite
{cmd} auth logout

# get auth status
{cmd} auth status`),
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(authStatusCmd)
}
