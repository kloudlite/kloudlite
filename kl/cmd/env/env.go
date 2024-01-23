package env

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "env",
	Short: "env specific commands",
	Long: `Use this command to switch and list environments
Examples:
		# list environments accessible to you
      kl env list

    # switch to a different environment
      kl env switch
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "envs")
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(listCmd)
}
