package expose

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "expose",
	Short: "expose ports",
	Long:  "This command will add ports to your kl-config file",
}

func init() {
	Command.AddCommand(portsCmd)
	Command.AddCommand(syncCmd)
	portsCmd.Aliases = []string{"ports"}
}
