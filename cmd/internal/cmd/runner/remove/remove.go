package remove

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/constants"
)

var RemoveCommand = &cobra.Command{
	Use:   "remove",
	Short: "remove [ secret | config | mres | mount ] configuration from your" + constants.CMD_NAME + "-config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
}

func init() {
	RemoveCommand.AddCommand(removeConfigCommand)
	RemoveCommand.AddCommand(removeMresCommand)
	RemoveCommand.AddCommand(removeSecretCommand)
	RemoveCommand.AddCommand(removeMountCommand)
}
