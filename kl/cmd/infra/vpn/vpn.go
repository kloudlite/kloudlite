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
  sudo kl infra vpn start

	# stop vpn
	sudo kl infra vpn stop

	# restart vpn
	sudo kl infra vpn restart

	# status vpn
	sudo kl infra vpn status
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(activateCmd)
	Cmd.AddCommand(intercept.Cmd)
}
