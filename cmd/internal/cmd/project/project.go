package project

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "project",
	Short: "list | use project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(dockerCredentialsCmd)
}
