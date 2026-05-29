package cmd

import "github.com/spf13/cobra"

func newServerCommand() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run Kloudlite server processes",
	}

	serverCmd.AddCommand(newAPIServerCommand())
	serverCmd.AddCommand(newWorkMachineManagerCommand())
	serverCmd.AddCommand(newTunnelServerCommand())
	return serverCmd
}
