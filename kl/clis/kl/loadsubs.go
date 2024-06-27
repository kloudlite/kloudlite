package kl

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/box"
	"github.com/kloudlite/kl/cmd/expose"
	"github.com/kloudlite/kl/cmd/get"
	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/packages"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"
	set_base_url "github.com/kloudlite/kl/cmd/set-base-url"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/vpn/intercept"
	"github.com/kloudlite/kl/flags"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	if flags.IsDev() {
		rootCmd.AddCommand(DocsCmd)
	}

	rootCmd.AddCommand(UpdateCmd)
	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(get.Cmd)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(box.BoxCmd)

	// apploader.LoadStartApp(rootCmd)

	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(set_base_url.Cmd)

	rootCmd.AddCommand(intercept.Cmd)

	rootCmd.AddCommand(expose.Command)

	rootCmd.AddCommand(add.Command)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(packages.Cmd)
}
