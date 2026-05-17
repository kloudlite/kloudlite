package cmd

import (
	"github.com/kloudlite/kloudlite/api/tunnel"
	"github.com/spf13/cobra"
)

func newTunnelServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "tunnel",
		Short:              "Run the tunnel server",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tunnel.Run(cmd.Context(), args)
		},
	}
}

func newTunnelServerCommandWithRunner(runner serverRunner) *cobra.Command {
	return &cobra.Command{
		Use:                "tunnel",
		Short:              "Run the tunnel server",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner(cmd.Context())
		},
	}
}
