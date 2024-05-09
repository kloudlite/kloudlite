package box

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/adrg/xdg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "get ssh access to the container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := sshBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func sshBox(cmd *cobra.Command, _ []string) error {
	debug := fn.ParseBoolFlag(cmd, "debug")
	command := exec.Command("ssh", "kl@localhost", "-p", "1729", "-i", path.Join(xdg.Home, ".ssh", "id_rsa"))

	if debug {
		fn.Log(command.String())
	}

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		fn.PrintError(fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running")))
		return err
	}
	return nil
}

func init() {
	sshCmd.Flags().BoolP("debug", "d", false, "run in debug mode")
}
