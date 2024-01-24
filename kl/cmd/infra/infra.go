package infra

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/cmd/infra/cluster"
	"github.com/kloudlite/kl/cmd/infra/vpn"

	"github.com/kloudlite/kl/cmd/infra/context"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:                "infra",
	Short:              "infra releated commands",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 || args[0] != "--" {
			if err := cmd.Help(); err != nil {
				functions.PrintError(err)
			}
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

			functions.Log(
				text.Bold(text.Green("\nSelected Cluster: ")),
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
			}, args[1:]); err != nil {
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
			functions.Log(fmt.Sprintf("%s=%s", k, v))
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
	Cmd.Aliases = append(Cmd.Aliases, "i")

	Cmd.Flags().StringP("cluster", "o", "", "cluster name")
	Cmd.Flags().StringP("account", "a", "", "account name")

	Cmd.AddCommand(context.Cmd)
	Cmd.AddCommand(vpn.Cmd)
	Cmd.AddCommand(cluster.Cmd)
}
