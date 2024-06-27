package use

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "env",
	Short: "Switch to a different environment",
	Run: func(cmd *cobra.Command, _ []string) {
		//TODO: add changes to the klbox-hash file
		envName := fn.ParseStringFlag(cmd, "envname")

		klFile, err := client.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		env, err := server.SelectEnv(envName, []fn.Option{
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if klFile.DefaultEnv == "" {
			klFile.DefaultEnv = env.Metadata.Name
			if err := client.WriteKLFile(*klFile); err != nil {
				fn.PrintError(err)
				return
			}
		}
		fn.Log(text.Bold(text.Green("\nSelected Environment:")),
			text.Blue(fmt.Sprintf("\n%s (%s)", env.DisplayName, env.Metadata.Name)),
		)

		wpath, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := hashctrl.SyncBoxHash(wpath); err != nil {
			return
		}
	},
}

func init() {
	switchCmd.Aliases = append(switchCmd.Aliases, "switch")

	switchCmd.Flags().StringP("envname", "e", "", "environment name")
	switchCmd.Flags().StringP("account", "a", "", "account name")
}
