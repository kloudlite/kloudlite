package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List [teams | envs | configs | secrets | apps]",
	Long: `Use this command to list resources like,
  team, environments, configs, secrets and apps`,
}

var InfraCmd = &cobra.Command{
	Use:   "list",
	Short: "List [teams | cluster]",
	Long:  `Use this command to list resources like, teams & clusters`,
}

func init() {
	Cmd.AddCommand(configsCmd)
	Cmd.AddCommand(secretsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(envCmd)

	Cmd.AddCommand(mresCmd)

	Cmd.Aliases = append(Cmd.Aliases, "ls")

	InfraCmd.AddCommand(accCmd)
	InfraCmd.Aliases = append(InfraCmd.Aliases, "ls")

	Cmd.PersistentFlags().StringP("output", "o", "table", "output format [table | json | yaml]")
}
