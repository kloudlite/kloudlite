package box

import (
	"fmt"
	"os"
	"os/exec"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec running container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := execBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func execBox(cmd *cobra.Command, args []string) error {

	cr, err := getRunningContainer()
	if err != nil {
		return err
	}

	if cr.Name == "" {
		return fmt.Errorf("no running container found")
	}

	debug := fn.ParseBoolFlag(cmd, "debug")
	command := exec.Command("docker", "exec", "-it", cr.Name, "bash")

	if debug {
		fn.Log(command.String())
	}

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		fn.PrintError(fmt.Errorf("failed to run command: %s", err))
		return err
	}
	return nil
}

func init() {
	execCmd.Flags().BoolP("debug", "d", false, "run in debug mode")
}
