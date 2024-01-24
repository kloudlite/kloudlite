package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "accounts | projects | env | configs | secrets | apps",
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

}
