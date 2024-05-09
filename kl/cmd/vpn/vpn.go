package vpn

import (
	"github.com/kloudlite/kl/cmd/vpn/intercept"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Example = functions.Desc(`# expose port of selected device
{cmd} vpn expose port -p <port>:<your_local_port>

# delete exposed port of selected device
{cmd} vpn expose port -d -p <port>:<your_local_port> 

# start vpn
sudo {cmd} vpn start

# stop vpn
sudo {cmd} vpn stop

# status vpn
sudo {cmd} vpn status
	  `)

var Cmd = &cobra.Command{
	Use:     "vpn",
	Short:   "vpn related commands",
	Example: Example,
	Long: `vpn related commands
Examples:
	`,
}

var InfraCmd = &cobra.Command{
	Use:     "vpn",
	Short:   "vpn related commands",
	Example: Example,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(startFgCmd)
	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(intercept.Cmd)

	InfraCmd.Aliases = append(InfraCmd.Aliases, "dev")
	InfraCmd.AddCommand(startCmd)
	InfraCmd.AddCommand(startFgCmd)
	InfraCmd.AddCommand(restartCmd)
	InfraCmd.AddCommand(stopCmd)
	InfraCmd.AddCommand(statusCmd)
	// InfraCmd.AddCommand(intercept.Cmd)
}
