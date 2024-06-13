package app

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden: true,
	Use:    "app",
	Short:  "app commands to start and stop controller app",
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(dnsCmd)
}
