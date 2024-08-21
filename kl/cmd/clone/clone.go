package clone

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "clone",
	Short: "clone environment from current environment",
}

func init() {
	Cmd.AddCommand(cloneCmd)
}
