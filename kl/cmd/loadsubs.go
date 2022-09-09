package cmd

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/create"
	"github.com/kloudlite/kl/cmd/get"
	"github.com/kloudlite/kl/cmd/intercept"
	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"
	"github.com/kloudlite/kl/cmd/runner/del"
	"github.com/kloudlite/kl/cmd/runner/gen"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/wg"
)

func init() {

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(get.Cmd)

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(wg.Cmd)

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

	rootCmd.AddCommand(create.Cmd)
}
