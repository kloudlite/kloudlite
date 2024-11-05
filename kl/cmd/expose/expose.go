package expose

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "expose",
	Short: "expose ports",
	Long:  "This command will add ports to your kl-config file",
}

func init() {
	Cmd.AddCommand(portsCmd)
	Cmd.AddCommand(syncCmd)
	portsCmd.Aliases = []string{"ports"}
}
