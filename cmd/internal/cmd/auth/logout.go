package auth

import (
	"fmt"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/lib"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := lib.Logout()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Successfully logged out")
	},
}
