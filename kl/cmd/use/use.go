package use

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "use",
	Short: "select env and account",
}

var InfraCmd = &cobra.Command{
	Use:   "use",
	Short: "select cluster and account",
}

func init() {
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(switchCmd)

	InfraCmd.AddCommand(accCmd)
	InfraCmd.AddCommand(clusterCmd)
}
