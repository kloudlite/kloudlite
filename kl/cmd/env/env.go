package env

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "env",
	Short: "env specific commands",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "envs")
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(listCmd)
}
