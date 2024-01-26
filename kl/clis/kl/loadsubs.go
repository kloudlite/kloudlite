package kl

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/get"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/spf13/cobra"

	// "github.com/kloudlite/kl/cmd/infra"
	sw "github.com/kloudlite/kl/cmd/switch"
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

	rootCmd.AddCommand(DocsCmd)
	rootCmd.AddCommand(UpdateCmd)

	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(get.Cmd)

	rootCmd.AddCommand(auth.Cmd)

	// rootCmd.AddCommand(infra.Cmd)
	rootCmd.AddCommand(vpn.Cmd)

	rootCmd.AddCommand(sw.Cmd)
	rootCmd.AddCommand(runner.InitCommand)

	rootCmd.AddCommand(add.Command)

	rootCmd.AddCommand(status.Cmd)
}
