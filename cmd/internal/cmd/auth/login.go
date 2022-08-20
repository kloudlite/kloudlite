package auth

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/cmd/account"
	"kloudlite.io/cmd/internal/lib"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "loging to kloudlite",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		lib.Login()
		account.TriggerSelectAccount()
	},
}
