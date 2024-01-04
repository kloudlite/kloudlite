package wg

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "wg",
	Short: "[connect | disconnect | reconnect | show] to wireguard service",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
}

func init() {
	Cmd.AddCommand(connectCmd)
	Cmd.AddCommand(disconnectCmd)
	Cmd.AddCommand(reconnectCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(exposeCmd)
}
