package app

import (
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop the kloudlite controller app",
	Long:  `This is internal command`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := Stop(); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("app stopped successfully")
	},
}

func Stop() error {
	p, err := proxy.NewProxy(true)
	if err != nil {
		return err
	}

	if err := p.Exit(); err != nil {
		return err
	}

	return nil
}
