package kli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                flags.CliName,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		if (len(args) != 0) && (args[0] == "--version" || args[0] == "-v") {
			fn.Log(cmd.Version)
			return
		}

		if len(args) < 2 || args[0] != "--" {
			if err := cmd.Help(); err != nil {
				fn.PrintError(err)
			}
			return
		}

		var fnn func() error
		fnn = func() error {

			accountName := fn.ParseStringFlag(cmd, "account")
			clusterName := fn.ParseStringFlag(cmd, "cluster")

			var err error
			clusterName, err = server.EnsureCluster([]fn.Option{
				fn.MakeOption("accountName", accountName),
				fn.MakeOption("clusterName", clusterName),
			}...)

			if err != nil {
				return err
			}

			fn.Log(
				text.Bold(text.Green("\nSelected Cluster: ")),
				text.Blue(fmt.Sprintf("%s", clusterName)),
			)

			configPath, err := server.SyncKubeConfig([]fn.Option{
				fn.MakeOption("accountName", accountName),
				fn.MakeOption("clusterName", clusterName),
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
		if err := fnn(); err != nil {
			fn.PrintError(err)
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
			fn.Log(fmt.Sprintf("%s=%s", k, v))
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if len(args) == 0 {
		return nil
	}

	return cmd.Run()
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = flags.Version
}
