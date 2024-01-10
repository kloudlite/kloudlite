package cmd

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/cluster"
	"github.com/kloudlite/kl/cmd/vpn"
	"github.com/spf13/cobra"

	"github.com/kloudlite/kl/cmd/context"
	"github.com/kloudlite/kl/cmd/get"
	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"

	"github.com/kloudlite/kl/cmd/env"
	"github.com/kloudlite/kl/cmd/runner/gen"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// hide help command

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	rootCmd.AddCommand(DocsCmd)

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(env.Cmd)
	rootCmd.AddCommand(get.Cmd)

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(context.Cmd)
	rootCmd.AddCommand(vpn.Cmd)
	//rootCmd.AddCommand(auth.logoutCmd)
	//rootCmd.AddCommand(auth.WhoAmICmd)

	rootCmd.AddCommand(cluster.Command)

	rootCmd.AddCommand(runner.InitCommand)

	// rootCmd.AddCommand(intercept.Cmd)

	rootCmd.AddCommand(add.Command)
	// rootCmd.AddCommand(del.DeleteCommand)
	rootCmd.AddCommand(gen.MountCommand)
}
