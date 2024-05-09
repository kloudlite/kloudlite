package use

import (
	"fmt"
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

		envName := fn.ParseStringFlag(cmd, "envname")
		env, err := server.SelectEnv(envName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		klFile, err := client.GetKlFile("")
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
	},
}

func init() {

	switchCmd.Aliases = append(switchCmd.Aliases, "switch")

	switchCmd.Flags().StringP("envname", "e", "", "environment name")
	switchCmd.Flags().StringP("projectname", "p", "", "project name")
	switchCmd.Flags().StringP("account", "a", "", "account name")
}
