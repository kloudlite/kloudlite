package list

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List [accounts | envs | configs | secrets | apps]",
	Long: `Use this command to list resources like,
  account, environments, configs, secrets and apps`,
}

var InfraCmd = &cobra.Command{
	Use:   "list",
	Short: "List [accounts | cluster]",
	Long:  `Use this command to list resources like, accounts & clusters`,
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
}
