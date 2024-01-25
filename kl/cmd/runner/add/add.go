package add

import (
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "import",
	Short: "import [ secret | config | mres | mount ] configuration to your kl-config file",
	Long:  `imports an environment variable from secret,config or managed resource`,
}

func init() {
	Command.AddCommand(confCmd)
	Command.AddCommand(mresCmd)
	Command.AddCommand(secCmd)
	Command.AddCommand(secCmd)
}
