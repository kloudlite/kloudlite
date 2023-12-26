package cluster

import (
	"fmt"
	util "github.com/kloudlite/kl/cmd/util"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "get access of your cluster",
	Long: `This command let you login to the kloudlite.
Example:
  # get detail about selected account
  kl cluster


  # exec new shell with kubeconfig env
  kl cluster -- bash

  # exec any kubernetes command
  kl cluster -- k9s
  kl cluster -- kubectl get pods
  kl cluster -- kubectl apply -f deployment.yaml

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var fn func()
		fn = func() {

			accountName := cmd.Flag("account").Value.String()
			clusterName := cmd.Flag("cluster").Value.String()

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
				switch err.Error() {
				case "noSelectedAccount":
					_, err := util.SelectAccount([]string{})
					if err != nil {
						fmt.Println(err)
						return
					}
					fn()
				case "noSelectedCluster":
					_, err := util.SelectCluster([]string{})
					if err != nil {
						fmt.Println(err)
						return
					}
					fn()
				default:
					fmt.Println(err)
				}
				return
			}
			if err := run(map[string]string{
				"KUBECONFIG": *configPath,
			}, args); err != nil {
				fmt.Println(err)
			}
		}
		fn()
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
