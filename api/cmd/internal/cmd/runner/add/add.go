package add

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/constants"
)

var AddCommand = &cobra.Command{
	Use:   "add",
	Short: "add [ secret | config | mres ] configuration to your " + constants.CMD_NAME + "-config file",
	Long:  `add an environment variable from secret,config or managed resource`,
}

func init() {
	AddCommand.AddCommand(addConfigCommand)
	AddCommand.AddCommand(addMresCommand)
	AddCommand.AddCommand(addSecretCommand)
	// AddCommand.AddCommand(addMountCommand)
}
