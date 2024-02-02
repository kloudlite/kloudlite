package add

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add",
	Short: "Add environment resources to your kl-config file",
	Long:  "Add secrets, configs, managed-resources and config-mounts to your kl-config file",
}

func init() {
	Command.AddCommand(confCmd)
	Command.AddCommand(mresCmd)
	Command.AddCommand(secCmd)
	Command.AddCommand(mountCommand)
}
