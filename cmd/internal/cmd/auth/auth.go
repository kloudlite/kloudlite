package auth

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login | logout to kloudlite",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
}
