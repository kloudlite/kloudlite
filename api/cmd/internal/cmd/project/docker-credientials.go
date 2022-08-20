package project

import (
	"fmt"
	"github.com/spf13/cobra"
)

var dockerCredentialsCmd = &cobra.Command{
	Use:   "docker-login",
	Short: "login to docker using the credential of selected project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}
