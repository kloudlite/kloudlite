package add

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add",
	Short: "add [ secret | config | mres ] configuration to your kl-config file",
	Long:  `add an environment variable from secret,config or managed resource`,
}

func init() {
	Command.AddCommand(addConfigCommand)
	Command.AddCommand(addMresCommand)
	Command.AddCommand(addSecretCommand)
	// Command.Command(addMountCommand)
}
