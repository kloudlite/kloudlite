package box

import (
	"errors"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "get ssh access to the container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := sshBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
		return
	},
}

func sshBox(_ *cobra.Command, _ []string) error {
	//containerName := "kl-box-" + getCwdHash()
	//command := exec.Command("docker", "exec", "-it", containerName, "bash")
	command := exec.Command("ssh", "kl@localhost", "-p", "1729")
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		fn.PrintError(errors.New(("Error opening ssh to kl-box container. Please ensure that container is running.")))
		return err
	}
	return nil
}

func init() {
	sshCmd.Aliases = append(sshCmd.Aliases, "ss")
}
