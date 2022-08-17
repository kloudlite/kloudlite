package cmd

import (
	"kloudlite.io/cmd/internal/cmd/account"
	"kloudlite.io/cmd/internal/cmd/auth"
	"kloudlite.io/cmd/internal/cmd/runner"
	"kloudlite.io/cmd/internal/cmd/wg"
)

func init() {

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(wg.ConnectCmd)
	rootCmd.AddCommand(wg.DisconnectCmd)

	rootCmd.AddCommand(account.Cmd)

	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(runner.AddCommand)
	rootCmd.AddCommand(runner.LoadCommand)

}
