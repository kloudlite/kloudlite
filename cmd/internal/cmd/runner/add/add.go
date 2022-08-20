package add

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/constants"
)

var AddCommand = &cobra.Command{
	Use:   "add",
	Short: "add [ secret | config | mres | mount ] configuration to your" + constants.CMD_NAME + "-config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
}

func init() {
	AddCommand.AddCommand(addConfigCommand)
	AddCommand.AddCommand(addMresCommand)
	AddCommand.AddCommand(addSecretCommand)
	AddCommand.AddCommand(addMountCommand)
}
