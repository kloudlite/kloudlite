package kli

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/spf13/cobra"

	// "github.com/kloudlite/kl/cmd/infra"
	sw "github.com/kloudlite/kl/cmd/switch"
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
	rootCmd.AddCommand(sw.InfraCmd)
	rootCmd.AddCommand(status.Cmd)
}
