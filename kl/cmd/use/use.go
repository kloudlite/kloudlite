package use

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "use",
	Short: "select environment and account to current context",
}

var InfraCmd = &cobra.Command{
	Use:   "use",
	Short: "select cluster and account to current context",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "select")
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(switchCmd)

	InfraCmd.AddCommand(accCmd)
	InfraCmd.AddCommand(clusterCmd)
}
