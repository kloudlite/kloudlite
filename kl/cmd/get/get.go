package get

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "get config/secret entries",
}

func init() {
	Cmd.AddCommand(configCmd)
	Cmd.AddCommand(secretCmd)
}
