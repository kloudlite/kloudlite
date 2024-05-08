package box

import (
	"errors"
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop running container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopBox("", cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopBox(containerName string, _ *cobra.Command, _ []string) error {
	if containerName == "" {
		containerName = "kl-box-" + getCwdHash()
	}
	fn.Log("stopping container...")
	if err := fn.ExecCmd(fmt.Sprintf("docker stop %s", containerName), nil, false); err != nil {
		fn.PrintError(errors.New("Error stoping kl-box container"))
		return err
	}

	if err := fn.ExecCmd(fmt.Sprintf("docker rm %s", containerName), nil, false); err != nil {
		return err
	}
	fn.Log("stopped container")
	return nil
}

func init() {
	stopCmd.Aliases = append(stopCmd.Aliases, "st")
}
