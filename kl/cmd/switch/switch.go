package sw

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "switch",
	Short: "select env and account",
	Example: `# list accounts accessible to you
kl account list

# switch to a different account
kl account switch
`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "sw")
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(switchCmd)

	Cmd.GroupID = "ctx"
}
