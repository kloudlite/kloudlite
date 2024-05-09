package kl

import (
	"github.com/kloudlite/kl/cmd/box"
	"github.com/kloudlite/kl/cmd/completion"
	"github.com/kloudlite/kl/cmd/shell"
	"runtime"

	"github.com/kloudlite/kl/flags"

	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/get"
	set_base_url "github.com/kloudlite/kl/cmd/set-base-url"
	app "github.com/kloudlite/kl/cmd/start-app"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/kloudlite/kl/constants"
	"github.com/spf13/cobra"

	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/vpn"

	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"
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

	rootCmd.AddCommand(vpn.Cmd)

	rootCmd.AddCommand(box.BoxCmd)

	if runtime.GOOS != constants.RuntimeWindows {
		rootCmd.AddCommand(app.Cmd)
	}

	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(set_base_url.Cmd)

	rootCmd.AddCommand(add.Command)

	rootCmd.AddCommand(status.Cmd)

	rootCmd.AddCommand(shell.ShellCmd)

	rootCmd.AddCommand(completion.AutoCompletion)

}
