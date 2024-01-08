package cluster

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "get access of your cluster",
	Long: `This command will let you perform different actions on your cluster.
Example:
  # get detail about selected account
  kl cluster


  # exec new shell with kubeconfig env
  kl cluster -- bash

  # exec any kubernetes command
  kl cluster -- k9s
  kl cluster -- kubectl get pods
  kl cluster -- kubectl apply -f deployment.yaml

	`,
	Run: func(cmd *cobra.Command, args []string) {
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

			accountName, err = client.CurrentAccountName()
			if err != nil {
				return err
			}

			configPath, err := server.SyncKubeConfig(func() *string {
				if accountName == "" {
					return nil
				}
				return &accountName
			}(), func() *string {
				if clusterName == "" {
					return nil
				}
				return &clusterName
			}())
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
	Command.Flags().StringP("cluster", "o", "", "cluster name")
	Command.Flags().StringP("account", "a", "", "account name")
}
