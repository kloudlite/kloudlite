package app

import (
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "stop the kloudlite controller app",
	Long:  `This is internal command`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := Status(); err != nil {
			fn.PrintError(err)
		}
	},
}

func Status() error {
	p, err := proxy.NewProxy(true)
	if err != nil {
		return err
	}

	if p.Status() {
		fn.Log("app is running")
		return nil
	}

	fn.Log("app is not running")
	return nil
}
