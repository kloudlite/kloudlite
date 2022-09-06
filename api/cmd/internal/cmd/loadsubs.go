package cmd

import (
	"kloudlite.io/cmd/internal/cmd/auth"
	"kloudlite.io/cmd/internal/cmd/get"
	"kloudlite.io/cmd/internal/cmd/intercept"
	"kloudlite.io/cmd/internal/cmd/list"
	"kloudlite.io/cmd/internal/cmd/runner"
	"kloudlite.io/cmd/internal/cmd/runner/add"
	"kloudlite.io/cmd/internal/cmd/runner/del"
	"kloudlite.io/cmd/internal/cmd/runner/gen"
	"kloudlite.io/cmd/internal/cmd/use"
)

func init() {

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(get.Cmd)

	rootCmd.AddCommand(auth.Cmd)

	// rootCmd.AddCommand(wg.ConnectCmd)
	// rootCmd.AddCommand(wg.DisconnectCmd)

	// rootCmd.AddCommand(account.Cmd)
	// rootCmd.AddCommand(project.Cmd)

	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(runner.LoadCommand)
	rootCmd.AddCommand(runner.ShowCommand)

	rootCmd.AddCommand(intercept.Cmd)
	rootCmd.AddCommand(intercept.LeaveCmd)

	rootCmd.AddCommand(add.AddCommand)
	rootCmd.AddCommand(del.DeleteCommand)
	rootCmd.AddCommand(gen.GenMountCommand)

}
