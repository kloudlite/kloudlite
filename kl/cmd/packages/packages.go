package packages

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pkg",
	Short: "packages util to manage nix packages of kl box",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "packages")
	Cmd.Aliases = append(Cmd.Aliases, "package")

	//fileclient.OnlyInsideBox(listCmd)
	//fileclient.OnlyInsideBox(addCmd)
	//fileclient.OnlyInsideBox(rmCmd)

	Cmd.AddCommand(listCmd)
	fileclient.OnlyInsideBox(addCmd)
	Cmd.AddCommand(addCmd)
	fileclient.OnlyInsideBox(rmCmd)
	Cmd.AddCommand(rmCmd)
	fileclient.OnlyInsideBox(searchCmd)
	Cmd.AddCommand(searchCmd)
}
