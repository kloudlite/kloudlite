package create

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "create new [device]",
}

func init() {
	Cmd.AddCommand(deviceCmd)
}
