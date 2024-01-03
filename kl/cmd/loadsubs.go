package cmd

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/cluster"
	// "github.com/kloudlite/kl/cmd/create"
	"github.com/kloudlite/kl/cmd/get"
	// "github.com/kloudlite/kl/cmd/intercept"
	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"
	// "github.com/kloudlite/kl/cmd/runner/del"
	"github.com/kloudlite/kl/cmd/runner/gen"
	switch_cmd "github.com/kloudlite/kl/cmd/switch"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/wg"
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(DocsCmd)

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(switch_cmd.Cmd)
	rootCmd.AddCommand(get.Cmd)

	rootCmd.AddCommand(auth.LoginCmd)
	rootCmd.AddCommand(auth.LogoutCmd)
	rootCmd.AddCommand(auth.WhoAmICmd)

	rootCmd.AddCommand(cluster.Command)

	rootCmd.AddCommand(wg.Cmd)

	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(runner.LoadCommand)
	rootCmd.AddCommand(runner.ShowCommand)

	// rootCmd.AddCommand(intercept.Cmd)
	// rootCmd.AddCommand(intercept.LeaveCmd)

	rootCmd.AddCommand(add.Command)
	// rootCmd.AddCommand(del.DeleteCommand)
	rootCmd.AddCommand(gen.MountCommand)

	// rootCmd.AddCommand(create.Cmd)
}
