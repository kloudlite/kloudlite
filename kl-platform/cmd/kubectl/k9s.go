package kubectl

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var K9sCmd = &cobra.Command{
	Use:                "k9s",
	Short:              "k9s is a terminal UI for Kubernetes",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("k9s", args...)
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
