package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "accounts | projects | devices | configs | secrets | apps | lambdas | regions",
	Long: `Using this command you can list multiple resources.
`,
}

func init() {
	Cmd.AddCommand(accountsCmd)
	Cmd.AddCommand(clustersCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(devicesCmd)
	Cmd.AddCommand(configsCmd)
	Cmd.AddCommand(secretsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(lambdasCmd)

	Cmd.AddCommand(regionsCmd)
}
