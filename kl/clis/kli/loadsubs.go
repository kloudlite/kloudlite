package kli

import (
	"runtime"

	"github.com/kloudlite/kl/cmd/auth"
	set_base_url "github.com/kloudlite/kl/cmd/set-base-url"
	app "github.com/kloudlite/kl/cmd/start-app"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/kloudlite/kl/constants"
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

	rootCmd.AddCommand(DocsCmd)
	rootCmd.AddCommand(UpdateCmd)

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(list.InfraCmd)
	rootCmd.AddCommand(vpn.InfraCmd)
	rootCmd.AddCommand(use.InfraCmd)
	rootCmd.AddCommand(status.Cmd)

	if runtime.GOOS == constants.RuntimeLinux {
		rootCmd.AddCommand(app.Cmd)
	}

	rootCmd.AddCommand(set_base_url.Cmd)
}
