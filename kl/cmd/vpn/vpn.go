package vpn

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "vpn",
	Short: "vpn command",
	Long:  `start/stop vpn`,
}

func init() {
	Cmd.AddCommand(startCmd)
	startCmd.Aliases = append(startCmd.Aliases, "connect", "up")
	Cmd.AddCommand(stopCmd)
	stopCmd.Aliases = append(stopCmd.Aliases, "down", "disconnect")
}
