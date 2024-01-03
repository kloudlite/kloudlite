package switch_cmd

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "switch",
	Short: "switch to a different environment",
	Long: `Switch Environment
Examples:
  # switch to a different environment
  kl switch envs

  # switch to a different environment with environment name
  kl switch envs <environment_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "sw")
	Cmd.AddCommand(envCmd)
}
