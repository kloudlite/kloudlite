package cmd

import (
	"kloudlite.io/cmd/internal/cmd/account"
	"kloudlite.io/cmd/internal/cmd/auth"
	"kloudlite.io/cmd/internal/cmd/project"
	"kloudlite.io/cmd/internal/cmd/runner"
	"kloudlite.io/cmd/internal/cmd/runner/add"
	"kloudlite.io/cmd/internal/cmd/runner/remove"
	"kloudlite.io/cmd/internal/cmd/wg"
)

func init() {

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(wg.ConnectCmd)
	rootCmd.AddCommand(wg.DisconnectCmd)

	rootCmd.AddCommand(account.Cmd)
	rootCmd.AddCommand(project.Cmd)

	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(runner.LoadCommand)
	rootCmd.AddCommand(runner.ShowCommand)

	rootCmd.AddCommand(add.AddCommand)
	rootCmd.AddCommand(remove.RemoveCommand)
}
