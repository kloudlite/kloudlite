package kli

import (
	"github.com/kloudlite/kl/cmd/auth"

	// "github.com/kloudlite/kl/cmd/infra"
	"github.com/kloudlite/kl/cmd/infra/vpn"
	sw "github.com/kloudlite/kl/cmd/switch"

	"github.com/kloudlite/kl/cmd/list"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(DocsCmd)
	rootCmd.AddCommand(UpdateCmd)

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(list.InfraCmd)
	rootCmd.AddCommand(vpn.Cmd)
	rootCmd.AddCommand(sw.InfraCmd)
}
