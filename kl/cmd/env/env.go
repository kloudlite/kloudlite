package env

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "env",
	Short: "env specific commands",
}

func init() {
	Cmd.AddCommand(switchCmd)
}
