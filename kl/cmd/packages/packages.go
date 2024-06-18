package packages

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pkg",
	Short: "packages util to manage nix packages of kl box",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "packages")
	Cmd.Aliases = append(Cmd.Aliases, "package")

	//client.OnlyInsideBox(listCmd)
	//client.OnlyInsideBox(addCmd)
	//client.OnlyInsideBox(rmCmd)

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(rmCmd)
	Cmd.AddCommand(searchCmd)
}
