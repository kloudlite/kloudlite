package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "list [ accounts | projects | devices | configs | secrets | apps | lambdas | routers ]",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.
`,
}

func init() {
	Cmd.AddCommand(accountsCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(devicesCmd)
	Cmd.AddCommand(configsCmd)
	Cmd.AddCommand(secretsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(lambdasCmd)
}
