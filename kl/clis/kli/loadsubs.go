package kli

import (
	apploader "github.com/kloudlite/kl/clis/app-loader"
	"github.com/kloudlite/kl/cmd/completion"
	"github.com/kloudlite/kl/cmd/shell"

	"github.com/kloudlite/kl/flags"

	"github.com/kloudlite/kl/cmd/auth"
	set_base_url "github.com/kloudlite/kl/cmd/set-base-url"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/spf13/cobra"

	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/vpn"

	"github.com/kloudlite/kl/cmd/list"
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

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(list.InfraCmd)
	rootCmd.AddCommand(vpn.InfraCmd)
	rootCmd.AddCommand(use.InfraCmd)
	rootCmd.AddCommand(status.Cmd)

	apploader.LoadStartApp(rootCmd)

	rootCmd.AddCommand(set_base_url.Cmd)

	rootCmd.AddCommand(shell.ShellCmd)

	rootCmd.AddCommand(completion.AutoCompletion)
}
