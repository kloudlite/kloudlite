package use

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "use",
	Short: "select environment to current context",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "select")
	Cmd.AddCommand(switchCmd)
}
