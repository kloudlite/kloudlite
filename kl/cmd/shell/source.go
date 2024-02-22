package shell

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path"
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
		_, err = server.EnsureAccount([]fn.Option{
			fn.MakeOption("accountName", accountName),
		}...)
		if err != nil {
			return err
		}
		klfile, err := client.GetKlFile("")
		if err != nil {
			return err
		}

		envs, cmap, smap, err := server.GetLoadMaps()
		if err != nil {
			return err
		}

		mountfiles := map[string]string{}

		for _, fe := range klfile.FileMount.Mounts {
			pth := fe.Path
			if pth == "" {
				pth = fe.Key
			}

			if fe.Type == client.ConfigType {
				mountfiles[pth] = cmap[fe.Name][fe.Key].Value
			} else {
				mountfiles[pth] = smap[fe.Name][fe.Key].Value
			}
		}

		if err = mounter.Mount(mountfiles, klfile.FileMount.MountBasePath); err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		envs["KL_MOUNT_PATH"] = path.Join(cwd, klfile.FileMount.MountBasePath)

		for index, value := range envs {
			newEnv = append(newEnv, fmt.Sprintf("%s=%s", index, value))
		}
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

	}
	newCmd := exec.Command(shellName())
	newCmd.Env = newEnv

	newCmd.Stdin = os.Stdin
	newCmd.Stdout = os.Stdout
	newCmd.Stderr = os.Stderr
	if err := newCmd.Run(); err != nil {
		return err
	}
	return nil
}

func shellName() string {

	shell := os.Getenv("SHELL")
	if shell[len(shell)-4:] == constants.BashShell {
		return constants.BashShell
	} else if shell[len(shell)-4:] == constants.FishShell {
		return constants.FishShell
	} else if shell[len(shell)-3:] == constants.ZshShell {
		return constants.ZshShell
	}
	return shell
}

func init() {
	ShellCmd.Aliases = append(ShellCmd.Aliases, "s", "sh")
	ShellCmd.Flags().StringP("account", "a", "", "account name")
	ShellCmd.Flags().StringP("cluster", "c", "", "cluster name")

}
