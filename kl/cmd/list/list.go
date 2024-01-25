package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "list [accounts | projects | env | configs | secrets | apps]",
	Long: `Using this command you can list multiple resources.
`,
}

var InfraCmd = &cobra.Command{
	Use:   "list",
	Short: "list accounts | cluster",
	Long: `Using this command you can list multiple resources.
`,
}

func init() {
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(configsCmd)
	Cmd.AddCommand(secretsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(envCmd)

	Cmd.Aliases = append(Cmd.Aliases, "ls")

	InfraCmd.AddCommand(accCmd)
	InfraCmd.AddCommand(clusterCmd)
	InfraCmd.Aliases = append(InfraCmd.Aliases, "ls")
}
