package shell

import (
	"fmt"
	// "github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/constants"
	// "github.com/kloudlite/kl/domain/client"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/domain/server"
	domain_util "github.com/kloudlite/kl/domain/util"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"

	// "path"
	"strings"
)

var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "loading environment variables to current shell",
	Long: `This command will load default environment variables to the current shell
Example:
{cmd} shell
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := loadEnv(cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func loadEnv(cmd *cobra.Command) error {
	accountName := fn.ParseStringFlag(cmd, "account")
	clusterName := fn.ParseStringFlag(cmd, "cluster")

	newEnv := exec.Command("kli -- printenv").Environ()
	var err error
	switch flags.CliName {
	case constants.CoreCliName:
		shell, err := ShellName()
		if err != nil {
			return err
		}

		if err := domain_util.MountEnv([]string{"--", shell}); err != nil {
			return err
		}

		return nil
	case constants.InfraCliName:
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
		newEnv = append(newEnv, fmt.Sprintf("KUBECONFIG=%s", *configPath))

		shell, err := ShellName()
		if err != nil {
			return err
		}
		newCmd := exec.Command(shell)
		newCmd.Env = newEnv

		newCmd.Stdin = os.Stdin
		newCmd.Stdout = os.Stdout
		newCmd.Stderr = os.Stderr
		if err := newCmd.Run(); err != nil {
			return err
		}

	}
	return nil
}

func ShellName() (string, error) {

	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		return "", fmt.Errorf("shell not found")
	}
	if strings.Contains(shell, constants.BashShell) {
		return constants.BashShell, nil
	} else if strings.Contains(shell, constants.FishShell) {
		return constants.FishShell, nil
	} else if strings.Contains(shell, constants.ZshShell) {
		return constants.ZshShell, nil
	} else if strings.Contains(shell, constants.PowerShell) {
		return constants.PowerShell, nil
	}
	return "", fmt.Errorf("unsupported shell")
}

func init() {
	ShellCmd.Aliases = append(ShellCmd.Aliases, "s", "sh")
	ShellCmd.Flags().StringP("account", "a", "", "account name")
	ShellCmd.Flags().StringP("cluster", "c", "", "cluster name")
}
