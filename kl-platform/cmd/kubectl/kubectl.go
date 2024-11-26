package kubectl

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var KubectlCmd = &cobra.Command{
	Use:                "kubectl",
	Short:              "kubectl is a command line tool for controlling Kubernetes clusters",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("kubectl", args...)
		c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", "/etc/rancher/k3s/k3s.yaml"))

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			err = nil
			return
		}

	},
}

func init() {
}
