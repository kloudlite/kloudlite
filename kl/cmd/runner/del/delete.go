package del

import (
	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"
)

var DeleteCommand = &cobra.Command{
	Use:   "del",
	Short: "delete [ secret | config | mres ] configuration from your" + constants.CmdName + "-config file",
	Long:  `Using this command you can delete added environment from the config file so next it will not load to your environment.`,
}

func init() {
	DeleteCommand.AddCommand(deleteConfigCommand)
	DeleteCommand.AddCommand(deleteMresCommand)
	DeleteCommand.AddCommand(deleteSecretCommand)
	// DeleteCommand.AddCommand(deleteMountCommand)
}
