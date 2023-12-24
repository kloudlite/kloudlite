package add

import (
	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add",
	Short: "add [ secret | config | mres ] configuration to your " + constants.CmdName + "-config file",
	Long:  `add an environment variable from secret,config or managed resource`,
}

func init() {
	Command.AddCommand(addConfigCommand)
	Command.AddCommand(addMresCommand)
	Command.AddCommand(addSecretCommand)
	// Command.Command(addMountCommand)
}
