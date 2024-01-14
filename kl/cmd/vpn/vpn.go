package vpn

import (
	"github.com/kloudlite/kl/cmd/vpn/intercept"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "vpn",
	Short: "work with vpn",
	Long: `work with vpn
Examples:
	# start vpn
  sudo kl vpn start

	# stop vpn
	sudo kl vpn stop

	# restart vpn
	sudo kl vpn restart

	# status vpn
	sudo kl vpn status

	# list all vpn
	kl vpn list

	# switch to vpn
	kl vpn switch <vpn_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(newCmd)
	Cmd.AddCommand(exposeCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(startFgCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(activateCmd)
	Cmd.AddCommand(intercept.Cmd)
}
