package infra

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/cmd/infra/context"
	"github.com/kloudlite/kl/cmd/infra/vpn"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "infra",
	Short: "create new infra context and manage existing infra contexts",
	Long: `Create new infra context and manage infra existing contexts
Examples:
  # creating new context
  kl infra context new

  # list all contexts
  kl infra context list

  # switch to context
  kl infra context switch <context_name>

  # remove context
  kl infra context remove <context_name>
	`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			cmd.Help()
			return
		}

		var fn func() error
		fn = func() error {

			accountName := cmd.Flag("account").Value.String()
			clusterName := cmd.Flag("cluster").Value.String()

			var err error
			clusterName, err = server.EnsureCluster([]functions.Option{
				functions.MakeOption("accountName", accountName),
				functions.MakeOption("clusterName", clusterName),
			}...)

			if err != nil {
				return err
			}

			fmt.Println(
				text.Bold(text.Green("\nSelected Cluster:")),
				text.Blue(fmt.Sprintf("%s", clusterName)),
			)

			configPath, err := server.SyncKubeConfig([]functions.Option{
				functions.MakeOption("accountName", accountName),
				functions.MakeOption("clusterName", clusterName),
			}...)

			if err != nil {
				return err
			}
			if err := run(map[string]string{
				"KUBECONFIG": *configPath,
			}, args); err != nil {
				return err
			}
			return nil
		}
		if err := fn(); err != nil {
			functions.PrintError(err)
			return
		}

	},
}

func run(envs map[string]string, args []string) error {
	var cmd *exec.Cmd
	if len(args) > 0 {
		argsWithoutProg := args[1:]
		cmd = exec.Command(args[0], argsWithoutProg...)
	} else {
		cmd = exec.Command("printenv")
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if len(args) > 0 {
		cmd.Env = os.Environ()
	}

	for k, v := range envs {
		if len(args) == 0 {
			fmt.Printf("%s=%q\n", k, v)
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if len(args) == 0 {
		return nil
	}

	return cmd.Run()
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "infra")

	Cmd.Flags().StringP("cluster", "o", "", "cluster name")
	Cmd.Flags().StringP("account", "a", "", "account name")

	Cmd.AddCommand(context.Cmd)
	Cmd.AddCommand(vpn.Cmd)
}
