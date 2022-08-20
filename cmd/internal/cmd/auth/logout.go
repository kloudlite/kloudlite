package auth

import (
	"fmt"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := lib.Logout()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("we expired the current session but the mounted config files are still present on your computer, run 'kl show' to see all the mounted files")
		fmt.Println("Successfully logged out")
	},
}
