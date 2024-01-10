package env

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch to a different environment",
	Long: `Switch Environment
Examples:
  # switch to a different environment
  kl env switch

  # switch to a different environment with environment name
  kl env switch --name <env_name>
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		envName := ""

		if cmd.Flags().Changed("name") {
			envName, _ = cmd.Flags().GetString("name")
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
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")

	switchCmd.Flags().StringP("name", "n", "", "environment name")
}
