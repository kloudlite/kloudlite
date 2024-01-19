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
  # expose port of selected device
	kl vpn expose port -p <port>:<your_local_port>

  # delete exposed port of selected device
	kl vpn expose port -d -p <port>:<your_local_port> 
	
	# start vpn
  sudo kl vpn start

	# stop vpn
	sudo kl vpn stop

	# restart vpn
	sudo kl vpn restart

	# status vpn
	sudo kl vpn status
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")

	Cmd.AddCommand(exposeCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(startFgCmd)
	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(intercept.Cmd)
}
