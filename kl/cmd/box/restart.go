package box

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart running container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := restartBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
		return
	},
}

func restartBox(_ *cobra.Command, _ []string) error {
	containerName := "kl-box-" + getCwdHash()
	isRunning, err := isContainerRunning(containerName)
	if err != nil {
		return err
	}
	if isRunning {
		if err := stopBox("", nil, nil); err != nil {
			return err
		}
	}

	if err := startBox(nil, nil); err != nil {
		return err
	}
	return nil
}

func isContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	running := strings.TrimSpace(string(output))
	return running == "true", nil
}

func init() {
	restartCmd.Aliases = append(restartCmd.Aliases, "rs")
}
