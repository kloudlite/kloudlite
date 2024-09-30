package connect

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/k3s"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "connect",
	Short: "start the wireguard connection",
	Long:  "This command will start the wireguard connection",
	Run: func(_ *cobra.Command, _ []string) {
		if err := startWg(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startWg() error {
	k3sClient, err := k3s.NewClient()
	if err != nil {
		return err
	}

	return k3sClient.RestartWgProxyContainer()
}
