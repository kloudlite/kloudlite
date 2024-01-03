package switch_cmd

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "switch to a different environment",
	Long: `Switch Environment
Examples:
  # switch to a different environment
  kl switch envs

  # switch to a different environment with environment name
  kl switch envs <environment_name>

	`,

	Run: func(_ *cobra.Command, args []string) {
		envName := ""

		if len(args) >= 1 {
			envName = args[0]
		}

		env, err := server.SelectEnv(envName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println(text.Bold(text.Green("\nSelected Environment:")),
			text.Blue(fmt.Sprintf("%s (%s)", env.DisplayName, env.Metadata.Name)),
		)
	},
}

func init() {
	envCmd.Aliases = append(envCmd.Aliases, "envs")
	envCmd.Aliases = append(envCmd.Aliases, "environment")
}
